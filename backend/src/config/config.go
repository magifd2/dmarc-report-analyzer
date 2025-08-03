package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all application configurations.
type Config struct {
	Port           int
	DBPath         string
	JWTSecret      string
	DataDir        string
	IPGeoDBPath    string
	ImportIPDBFile string // Path to MMDB file for manual import via CLI

	// CLI options for user management
	CreateUserUsername string
	CreateUserPassword string
}

// LoadConfig loads configuration from command-line flags and environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Define command-line flags
	flag.IntVar(&cfg.Port, "port", 8080, "Port for the HTTP server to listen on")
	flag.StringVar(&cfg.DBPath, "db-path", "", "Path to the SQLite database file (default: data/dmarc_reports.db)")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", os.Getenv("JWT_SECRET"), "Secret key for JWT signing (environment variable JWT_SECRET)")
	flag.StringVar(&cfg.DataDir, "data-dir", "data", "Directory for application data (database, IP geo files)")
	flag.StringVar(&cfg.ImportIPDBFile, "import-ip-db", "", "Path to an IPInfo MMDB file to import (e.g., /path/to/ipinfo-city.mmdb)")

	// New flags for user creation
	flag.StringVar(&cfg.CreateUserUsername, "create-user", "", "Create a new user with the given username")
	flag.StringVar(&cfg.CreateUserPassword, "password", "", "Password for the new user (used with --create-user)")

	flag.Parse()

	// Determine the application root directory
	appRoot, err := GetAppRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to determine application root: %w", err)
	}

	// Set DataDir relative to the application root
	cfg.DataDir = filepath.Join(appRoot, cfg.DataDir)

	// Ensure DataDir exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", cfg.DataDir, err)
	}

	// Set default DBPath if not provided
	if cfg.DBPath == "" {
		cfg.DBPath = filepath.Join(cfg.DataDir, "dmarc_reports.db")
	} else {
		// If DBPath is provided, ensure it's absolute or relative to appRoot
		if !filepath.IsAbs(cfg.DBPath) {
			cfg.DBPath = filepath.Join(appRoot, cfg.DBPath)
		}
	}

	// Set IPGeoDBPath
	cfg.IPGeoDBPath = filepath.Join(cfg.DataDir, "ip_geo")
	if err := os.MkdirAll(cfg.IPGeoDBPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create IP geo database directory %s: %w", cfg.IPGeoDBPath, err)
	}

	// Validate JWT Secret only if not creating a user or importing IP DB
	// If creating a user or importing IP DB, we might not need the server to run
	if cfg.CreateUserUsername == "" && cfg.ImportIPDBFile == "" {
		if cfg.JWTSecret == "" {
			return nil, fmt.Errorf("JWT_SECRET environment variable is not set. This is required for authentication.")
		}
		if len(cfg.JWTSecret) < 32 { // Recommend a reasonably long secret
			fmt.Println("Warning: JWT_SECRET is too short. It should be at least 32 characters for security.")
		}
	}

	return cfg, nil
}

// GetAppRoot returns the absolute path to the application's root directory.
// This assumes the executable is directly in the root or a known subdirectory.
func GetAppRoot() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	// Get the directory of the executable
	execDir := filepath.Dir(ex)

	// If running from `backend/` during development, adjust path
	// This is a heuristic and might need adjustment based on build/run environment
	if strings.HasSuffix(execDir, filepath.Join("backend", "bin")) { // When built by start.sh
		return filepath.Dir(filepath.Dir(execDir)), nil // Go up two levels from backend/bin
	} else if strings.HasSuffix(execDir, filepath.Join("backend", "src")) { // When run directly from src
		return filepath.Dir(filepath.Dir(execDir)), nil // Go up two levels from backend/src
	} else if strings.HasSuffix(execDir, "backend") { // When run directly from backend
		return filepath.Dir(execDir), nil // Go up one level from backend
	}
	return execDir, nil // Assume executable is in the root for packaged app
}
