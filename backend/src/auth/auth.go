package auth

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"dmarc-report-analyzer/backend/src/db"
)

// Claims defines the JWT claims structure.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthService provides authentication related functionalities.
type AuthService struct {
	DBRepo *db.Repository
	JWTSecret []byte
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(dbRepo *db.Repository, jwtSecret string) *AuthService {
	return &AuthService{
		DBRepo: dbRepo,
		JWTSecret: []byte(jwtSecret),
	}
}

// HashPassword hashes a plain text password using bcrypt.
func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash compares a plain text password with a hashed password.
func (s *AuthService) CheckPasswordHash(password, hash string) bool {
	log.Printf("Comparing password (plain): %s with hash: %s", password, hash)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Printf("Password comparison failed: %v", err)
	}
	return err == nil
}

// GenerateJWT generates a new JWT token for a given username.
func (s *AuthService) GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "dmarc-report-analyzer",
			Subject:   username,
			Audience:  []string{"users"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims if valid.
func (s *AuthService) ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.JWTSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}