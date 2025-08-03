package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"dmarc-report-analyzer/backend/src/auth"
	"dmarc-report-analyzer/backend/src/db"
)

// UsersAPI handles user related API endpoints.
type UsersAPI struct {
	AuthService *auth.AuthService
	DBRepo      *db.Repository
}

// NewUsersAPI creates a new UsersAPI instance.
func NewUsersAPI(authService *auth.AuthService, dbRepo *db.Repository) *UsersAPI {
	return &UsersAPI{
		AuthService: authService,
		DBRepo:      dbRepo,
	}
}

// RegisterUserRoutes registers the user API routes.
func RegisterUserRoutes(router *mux.Router, api *UsersAPI) {
	// This route should be protected by authentication middleware
	router.HandleFunc("/api/users/change-password", api.ChangePassword).Methods("POST")
}

// ChangePassword handles changing a user's own password.
func (api *UsersAPI) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement authentication middleware to get the authenticated user's ID/username
	// For now, we'll assume the username is passed in the request body for simplicity

	var req struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := api.DBRepo.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("Error getting user %s: %v", req.Username, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil || !api.AuthService.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		http.Error(w, "Invalid username or old password", http.StatusUnauthorized)
		return
	}

	hashedNewPassword, err := api.AuthService.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing new password for user %s: %v", req.Username, err)
		http.Error(w, "Failed to change password", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = hashedNewPassword
	err = api.DBRepo.UpdateUser(user)
	if err != nil {
		log.Printf("Error updating password for user %s: %v", req.Username, err)
		http.Error(w, "Failed to change password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
}
