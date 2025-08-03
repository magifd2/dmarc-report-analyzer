package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"

	"dmarc-report-analyzer/backend/src/api"
	"dmarc-report-analyzer/backend/src/config"
	"dmarc-report-analyzer/backend/src/core/parser"
	"dmarc-report-analyzer/backend/src/db"
	"dmarc-report-analyzer/backend/src/ip_geo"
)

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

	// 3. Initialize IP Geo Resolver
	ipResolver, err := ip_geo.NewResolver(cfg.IPGeoDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize IP geo resolver: %v", err)
	}
	// If IP geo databases were not loaded, log a warning
	if ipResolver.CityDB == nil && ipResolver.AsnDB == nil {
		log.Println("Warning: IP Geo databases are not loaded. IP resolution will not be available.")
	}

	// 4. Setup HTTP Router
	router := mux.NewRouter()

	// Initialize Parser and ReportsAPI
	reportProcessor := parser.NewReportProcessor(dbRepo, ipResolver)
	reportsAPI := api.NewReportsAPI(reportProcessor, dbRepo)

	// Register API routes
	api.RegisterReportRoutes(router, reportsAPI)
	// TODO: Register Auth and Settings routes

	// Serve static files (frontend)
	staticPath := filepath.Join(cfg.DataDir, "..", "static") // Adjust path as needed
	if _, err := os.Stat(staticPath); os.IsNotExist(err) {
		log.Printf("Warning: Static files directory not found at %s. Frontend may not be served.", staticPath)
	}
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(staticPath)))


	// 5. Start HTTP Server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
