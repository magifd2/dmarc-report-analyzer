package db

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user in the database.
func (r *Repository) CreateUser(username, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now().Unix(),
	}

	stmt, err := r.db.Prepare(`
		INSERT INTO users (username, password_hash, created_at)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement for creating user: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(user.Username, user.PasswordHash, user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for creating user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID for user: %w", err)
	}
	user.ID = id
	return user, nil
}

// GetUserByUsername retrieves a user by their username.
func (r *Repository) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow("SELECT id, username, password_hash, created_at FROM users WHERE username = ?", username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to query user by username: %w", err)
	}
	return user, nil
}

// UpdateUserPassword updates a user's password.
func (r *Repository) UpdateUserPassword(userID int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	stmt, err := r.db.Prepare("UPDATE users SET password_hash = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement for updating password: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(string(hashedPassword), userID)
	if err != nil {
		return fmt.Errorf("failed to execute statement for updating password: %w", err)
	}
	return nil
}

// DeleteUser deletes a user by their ID.
func (r *Repository) DeleteUser(userID int64) error {
	stmt, err := r.db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement for deleting user: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID)
	if err != nil {
		return fmt.Errorf("failed to execute statement for deleting user: %w", err)
	}
	return nil
}

// VerifyPassword verifies a plain-text password against a hashed password.
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
