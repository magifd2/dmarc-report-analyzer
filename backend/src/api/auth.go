package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"dmarc-report-analyzer/backend/src/auth"
	"dmarc-report-analyzer/backend/src/db"
)

// AuthAPI handles authentication related API endpoints.
type AuthAPI struct {
	AuthService *auth.AuthService
	DBRepo      *db.Repository
}

// NewAuthAPI creates a new AuthAPI instance.
func NewAuthAPI(authService *auth.AuthService, dbRepo *db.Repository) *AuthAPI {
	return &AuthAPI{
		AuthService: authService,
		DBRepo:      dbRepo,
	}
}

// RegisterAuthRoutes registers the authentication API routes.
func RegisterAuthRoutes(router *mux.Router, api *AuthAPI) {
	router.HandleFunc("/api/auth/login", api.Login).Methods("POST")
}

// Login handles user login and JWT generation.
func (api *AuthAPI) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt for user: %s", creds.Username)

	user, err := api.DBRepo.GetUserByUsername(creds.Username)
	if err != nil {
		log.Printf("Error getting user %s from DB: %v", creds.Username, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		log.Printf("User %s not found in DB.", creds.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("User %s found in DB. Stored hash: %s", creds.Username, user.PasswordHash)

	if !api.AuthService.CheckPasswordHash(creds.Password, user.PasswordHash) {
		log.Printf("Password mismatch for user: %s", creds.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Password match for user: %s. Generating JWT.", creds.Username)
	tokenString, err := api.AuthService.GenerateJWT(user.Username)
	if err != nil {
		log.Printf("Error generating JWT for user %s: %v", user.Username, err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"token": tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}