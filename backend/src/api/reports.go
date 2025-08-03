package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"dmarc-report-analyzer/backend/src/core/parser"
	"dmarc-report-analyzer/backend/src/db"
)

// ReportsAPI handles DMARC report related API endpoints.
type ReportsAPI struct {
	Processor *parser.ReportProcessor
	DBRepo    *db.Repository
}

// NewReportsAPI creates a new ReportsAPI instance.
func NewReportsAPI(processor *parser.ReportProcessor, dbRepo *db.Repository) *ReportsAPI {
	return &ReportsAPI{
		Processor: processor,
		DBRepo:    dbRepo,
	}
}

// RegisterReportRoutes registers the DMARC report API routes.
func RegisterReportRoutes(router *mux.Router, api *ReportsAPI) {
	router.HandleFunc("/api/reports/upload", api.UploadReports).Methods("POST")
	// TODO: Add other report related routes (e.g., /api/reports, /api/reports/{id}, /api/reports/summary)
}

// UploadReports handles the upload of DMARC aggregate reports.
func (api *ReportsAPI) UploadReports(w http.ResponseWriter, r *http.Request) {
	// Limit the size of the request body to prevent abuse
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100 MB limit

	reader, err := r.MultipartReader()
	if err != nil {
		if err.Error() == "http: request body too large" {
			http.Error(w, "Request body too large. Max 100MB.", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to read multipart form: %v", err), http.StatusBadRequest)
		return
	}

	var totalProcessed int
	var totalSkipped int
	var totalFailed int
	var fileErrors []map[string]string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break // All parts read
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read next part: %v", err), http.StatusInternalServerError)
			return
		}

		if part.FileName() == "" {
			continue // Skip non-file parts
		}

		log.Printf("Processing uploaded file: %s", part.FileName())

		// Process the file using the parser
		ingestionErrors := api.Processor.ProcessUploadedFile(part, part.FileName())

		// Iterate through ingestionErrors to correctly count processed, skipped, and failed
		if len(ingestionErrors) > 0 {
			for _, errInfo := range ingestionErrors {
				if errInfo.ErrorType == "SKIPPED_DUPLICATE" {
					totalSkipped++
				} else {
					totalFailed++
					fileErrors = append(fileErrors, map[string]string{
						"filename":   errInfo.Filename,
						"error_type": errInfo.ErrorType,
						"message":    errInfo.Message,
					})
				}
				// Save ingestion error to DB regardless of type (skipped or failed)
				if err := api.DBRepo.SaveIngestionError(&errInfo); err != nil {
					log.Printf("Failed to save ingestion error to DB: %v", err)
				}
			}
		} else {
			// If no ingestion errors, it means the file was processed successfully
			totalProcessed++
		}
	}

	response := map[string]interface{}{
		"status":           "success",
		"message":          "Reports processing completed.",
		"processed_count":  totalProcessed,
		"skipped_count":    totalSkipped,
		"failed_files_count": totalFailed,
		"file_errors":      fileErrors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
