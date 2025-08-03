package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// InitDB initializes the SQLite database and performs migrations.
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency and performance
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Printf("Warning: Failed to enable WAL mode: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return db, nil
}

// runMigrations applies necessary schema changes.
func runMigrations(db *sql.DB) error {
	// Version 1: Initial schema
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			xml_hash TEXT UNIQUE NOT NULL,
			original_xml TEXT NOT NULL,
			org_name TEXT,
			report_id TEXT,
			date_range_begin INTEGER,
			date_range_end INTEGER,
			domain TEXT,
			adkim TEXT,
			aspf TEXT,
			p TEXT,
			sp TEXT,
			pct INTEGER
		);

		CREATE TABLE IF NOT EXISTS records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			report_id INTEGER NOT NULL,
			source_ip TEXT NOT NULL,
			count INTEGER,
			header_from TEXT,
			disposition TEXT,
			dkim_result TEXT,
			spf_result TEXT,
			FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS ip_info (
			ip_address TEXT PRIMARY KEY,
			country_code TEXT,
			country_name TEXT,
			city_name TEXT,
			asn_number INTEGER,
			asn_organization TEXT,
			hostname TEXT,
			reversed_hostname TEXT,
			apex_domain TEXT,
			last_updated INTEGER
		);

		CREATE TABLE IF NOT EXISTS ingestion_errors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,
			xml_hash TEXT,
			error_type TEXT NOT NULL,
			message TEXT NOT NULL,
			timestamp INTEGER NOT NULL
		);

		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);

		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create initial schema (migration v1): %w", err)
	}

	// Add indexes for performance
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_reports_date_range_begin ON reports(date_range_begin);
		CREATE INDEX IF NOT EXISTS idx_records_report_id ON records(report_id);
		CREATE INDEX IF NOT EXISTS idx_records_source_ip ON records(source_ip);
		CREATE INDEX IF NOT EXISTS idx_records_header_from ON records(header_from);
		CREATE INDEX IF NOT EXISTS idx_ip_info_hostname ON ip_info(hostname);
		CREATE INDEX IF NOT EXISTS idx_ip_info_reversed_hostname ON ip_info(reversed_hostname);
		CREATE INDEX IF NOT EXISTS idx_ip_info_apex_domain ON ip_info(apex_domain);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Database schema initialized/migrated successfully.")
	return nil
}
