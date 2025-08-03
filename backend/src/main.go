package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"dmarc-report-analyzer/backend/src/api"
	"dmarc-report-analyzer/backend/src/auth"
	"dmarc-report-analyzer/backend/src/config"
	"dmarc-report-analyzer/backend/src/core/parser"
	"dmarc-report-analyzer/backend/src/db"
	"dmarc-report-analyzer/backend/src/ip_geo"
)

//go:embed static_frontend_dist/*
var embeddedFiles embed.FS

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Handle --import-ip-db CLI option
	if cfg.ImportIPDBFile != "" {
		log.Printf("Attempting to import IP database from: %s", cfg.ImportIPDBFile)
		if err := ip_geo.ImportMMDBFile(cfg.ImportIPDBFile, cfg.IPGeoDBPath); err != nil {
			log.Fatalf("Failed to import IP database: %v", err)
		}
		log.Println("IP database imported successfully. Exiting.")
		os.Exit(0) // Exit after import
	}

	// 2. Initialize Database
	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	dbRepo := db.NewRepository(database)

	// Handle --create-user CLI option
	if cfg.CreateUserUsername != "" {
		if cfg.CreateUserPassword == "" {
			log.Fatalf("Error: --password must be provided when using --create-user.")
		}
		log.Printf("Attempting to create user: %s", cfg.CreateUserUsername)
		_, err := dbRepo.CreateUser(cfg.CreateUserUsername, cfg.CreateUserPassword)
		if err != nil {
			log.Fatalf("Failed to create user %s: %v", cfg.CreateUserUsername, err)
		}
		log.Printf("User %s created successfully. Exiting.", cfg.CreateUserUsername)
		os.Exit(0) // Exit after user creation
	}

	// 3. Initialize IP Geo Resolver
	ipResolver, err := ip_geo.NewResolver(cfg.IPGeoDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize IP geo resolver: %v", err)
	}
	// If IP geo databases were not loaded, log a warning
	if ipResolver.CityDB == nil {
		log.Println("Warning: IP Geo databases are not loaded. IP resolution will not be available.")
	}

	// 4. Initialize Auth Service
	authService := auth.NewAuthService(dbRepo, cfg.JWTSecret)

	// 5. Setup HTTP Router
	router := mux.NewRouter()

	// Initialize API handlers
	reportProcessor := parser.NewReportProcessor(dbRepo, ipResolver)
	reportsAPI := api.NewReportsAPI(reportProcessor, dbRepo)
	authAPI := api.NewAuthAPI(authService, dbRepo)
	usersAPI := api.NewUsersAPI(authService, dbRepo)

	// Register API routes
	api.RegisterReportRoutes(router, reportsAPI)
	api.RegisterAuthRoutes(router, authAPI)
	api.RegisterUserRoutes(router, usersAPI)

	// static_frontend_dist サブディレクトリをルートとして扱う
	staticFiles, err := fs.Sub(embeddedFiles, "static_frontend_dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}

	// "/" を static_frontend_dist にマッピング
	router.PathPrefix("/").Handler(spaHandler(staticFiles))

	// Configure CORS middleware
	c := cors.AllowAll() // For development, allow all origins
	handler := c.Handler(router)

	// サーバー起動
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting at %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func spaHandler(staticFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(staticFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// URLからファイルパスを取得（先頭スラッシュ除去）
		path := strings.TrimPrefix(r.URL.Path, "/")

		// ファイルが存在するか確認
		_, err := fs.Stat(staticFS, path)
		if err == nil {
			// 実ファイルが存在 → 通常通り配信
			fileServer.ServeHTTP(w, r)
			return
		}

		// ファイルが存在しない → index.html or 404.html を返す（SPA対応）
		// まず 404.html があれば使う、なければ index.html を使う
		fallback := "index.html"
		if _, err := fs.Stat(staticFS, "404.html"); err == nil && !strings.HasPrefix(path, "api/") {
			fallback = "404.html"
		}

		data, err := fs.ReadFile(staticFS, fallback)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if fallback == "index.html" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
		} else {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
		}
		w.Write(data)
	})
}