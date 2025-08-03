package parser

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"log" // Consider replacing with a structured logger like zap or logrus
	"path/filepath"
	"strings"
	"time"

	"dmarc-report-analyzer/backend/src/db"
	"dmarc-report-analyzer/backend/src/ip_geo"
)

// ReportProcessor handles the parsing and storage of DMARC reports.
type ReportProcessor struct {
	DBRepo     *db.Repository
	IPResolver *ip_geo.Resolver
}

// NewReportProcessor creates a new ReportProcessor instance.
func NewReportProcessor(dbRepo *db.Repository, ipResolver *ip_geo.Resolver) *ReportProcessor {
	return &ReportProcessor{
		DBRepo:     dbRepo,
		IPResolver: ipResolver,
	}
}

// ProcessUploadedFile processes a single uploaded DMARC report file.
// It returns a list of ingestion errors for the file.
func (rp *ReportProcessor) ProcessUploadedFile(file io.Reader, filename string) []db.IngestionError {
	var ingestionErrors []db.IngestionError

	// 1. ファイルヘッダの読み込みと厳密なファイルタイプ識別
	// io.ReaderからPeekableReaderを作成し、先頭数KBを読み込む
	peekReader := newPeekableReader(file)
	header, err := peekReader.Peek(4096) // Read first 4KB
	if err != nil && err != io.EOF {
		ingestionErrors = append(ingestionErrors, db.IngestionError{
			Filename:  filename,
			ErrorType: "FILE_READ_ERROR",
			Message:   fmt.Sprintf("Failed to read file header: %v", err),
			Timestamp: time.Now().Unix(),
		})
		return ingestionErrors
	}

	fileType := identifyFileType(header, filename)
	if fileType == FileTypeUnknown {
		ingestionErrors = append(ingestionErrors, db.IngestionError{
			Filename:  filename,
			ErrorType: "UNSUPPORTED_FILE_TYPE",
			Message:   "File format not recognized as XML, ZIP, GZ, or TGZ.",
			Timestamp: time.Now().Unix(),
		})
		return ingestionErrors
	}

	// 2. アーカイブの展開 / XMLコンテンツの取得
	xmlContents, err := rp.extractXMLContent(peekReader, fileType)
	if err != nil {
		ingestionErrors = append(ingestionErrors, db.IngestionError{
			Filename:  filename,
			ErrorType: "ARCHIVE_EXTRACTION_ERROR",
			Message:   fmt.Sprintf("Failed to extract XML from archive: %v", err),
			Timestamp: time.Now().Unix(),
		})
		return ingestionErrors
	}

	// Process each extracted XML content (for ZIP/TGZ, there might be multiple)
	if len(xmlContents) == 0 {
		ingestionErrors = append(ingestionErrors, db.IngestionError{
			Filename:  filename,
			ErrorType: "NO_XML_CONTENT",
			Message:   "No XML content found in the provided file or archive.",
			Timestamp: time.Now().Unix(),
		})
		return ingestionErrors
	}

	for _, xmlContent := range xmlContents {
		reportErrors := rp.processSingleXML(xmlContent, filename)
		ingestionErrors = append(ingestionErrors, reportErrors...)
	}

	return ingestionErrors
}

// processSingleXML processes a single DMARC XML content.
func (rp *ReportProcessor) processSingleXML(xmlContent []byte, originalFilename string) []db.IngestionError {
	var ingestionErrors []db.IngestionError

	// 3. XMLハッシュの計算
	xmlHash := fmt.Sprintf("%x", sha256.Sum256(xmlContent))

	// 4. 重複チェック
	exists, err := rp.DBRepo.ReportExistsByHash(xmlHash)
	if err != nil {
		ingestionErrors = append(ingestionErrors, db.IngestionError{
			Filename:  originalFilename,
			XMLHash:   xmlHash,
			ErrorType: "DB_CHECK_ERROR",
			Message:   fmt.Sprintf("Failed to check for duplicate report: %v", err),
			Timestamp: time.Now().Unix(),
		})
		return ingestionErrors
	}
	if exists {
		log.Printf("Report with hash %s already exists. Skipping.", xmlHash)
		// Return a specific error type for skipped duplicates
		return []db.IngestionError{{
			Filename:  originalFilename,
			XMLHash:   xmlHash,
			ErrorType: "SKIPPED_DUPLICATE",
			Message:   "Report with this hash already exists. Skipped.",
			Timestamp: time.Now().Unix(),
		}}
	}

	// 5. XMLパースと厳密なバリデーション
	log.Printf("Raw XML content for unmarshalling:\n%s", string(xmlContent)) // Debugging line
	var feedback Feedback
	if err := xml.Unmarshal(xmlContent, &feedback); err != nil {
		ingestionErrors = append(ingestionErrors, db.IngestionError{
			Filename:  originalFilename,
			XMLHash:   xmlHash,
			ErrorType: "XML_PARSE_ERROR",
			Message:   fmt.Sprintf("Failed to parse XML: %v", err),
			Timestamp: time.Now().Unix(),
		})
		return ingestionErrors
	}

	// Perform strict DMARC schema validation
	if err := feedback.Validate(); err != nil {
		ingestionErrors = append(ingestionErrors, db.IngestionError{
			Filename:  originalFilename,
			XMLHash:   xmlHash,
			ErrorType: "DMARC_VALIDATION_ERROR",
			Message:   fmt.Sprintf("DMARC report validation failed: %v", err),
			Timestamp: time.Now().Unix(),
		})
		return ingestionErrors
	}

	// 6. IP情報解決
	// Collect unique IPs to resolve them efficiently
	uniqueIPs := make(map[string]struct{})
	for _, record := range feedback.Records {
	
uniqueIPs[record.Row.SourceIP] = struct{}{}
	}

	for ipStr := range uniqueIPs {
		ipInfo, err := rp.IPResolver.ResolveIP(ipStr)
		if err != nil {
			log.Printf("Failed to resolve IP %s: %v. Storing partial IPInfo.", ipStr, err)
			// Create a minimal IPInfo with "N/A" for unresolved fields
			ipInfo = &db.IPInfo{
				IPAddress: ipStr,
				Hostname: "N/A", ReversedHostname: "N/A", ApexDomain: "N/A",
				CountryCode: "N/A", CountryName: "N/A", CityName: "N/A",
				ASNNumber: 0, ASNOrganization: "N/A",
				LastUpdated: time.Now().Unix(),
			}
		}
		// Save or update IPInfo in DB
		if err := rp.DBRepo.SaveOrUpdateIPInfo(ipInfo); err != nil {
			log.Printf("Failed to save/update IPInfo for %s: %v", ipStr, err)
			// This error is logged but doesn't stop report ingestion
		}
	}

	// 7. データベースへの保存
	report := db.Report{
		XMLHash:        xmlHash,
		OriginalXML:    string(xmlContent),
		OrgName:        feedback.ReportMetadata.OrgName,
		ReportID:       feedback.ReportMetadata.ReportID,
		DateRangeBegin: feedback.ReportMetadata.DateRange.Begin,
		DateRangeEnd:   feedback.ReportMetadata.DateRange.End,
		Domain:         feedback.PolicyPublished.Domain,
		ADKIM:          feedback.PolicyPublished.ADKIM,
		ASPF:           feedback.PolicyPublished.ASPF,
		P:              feedback.PolicyPublished.P,
		SP:             feedback.PolicyPublished.SP,
		PCT:            feedback.PolicyPublished.PCT,
	}

	reportID, err := rp.DBRepo.SaveReport(&report)
	if err != nil {
		ingestionErrors = append(ingestionErrors, db.IngestionError{
			Filename:  originalFilename,
			XMLHash:   xmlHash,
			ErrorType: "DB_SAVE_ERROR",
			Message:   fmt.Sprintf("Failed to save report to database: %v", err),
			Timestamp: time.Now().Unix(),
		})
		return ingestionErrors
	}

	for _, record := range feedback.Records {
		dbRecord := db.Record{
			ReportID:    reportID,
			SourceIP:    record.Row.SourceIP,
			Count:       record.Row.Count,
			HeaderFrom:  record.Identifiers.HeaderFrom,
			Disposition: record.Row.PolicyEvaluated.Disposition, // Access through record.Row
			DKIMResult:  record.Row.PolicyEvaluated.DKIM,        // Access through record.Row
			SPFResult:   record.Row.PolicyEvaluated.SPF,         // Access through record.Row
		}
		if err := rp.DBRepo.SaveRecord(&dbRecord); err != nil {
			log.Printf("Failed to save record for report %d, IP %s: %v", reportID, record.Row.SourceIP, err)
			// This error is logged but doesn't stop report ingestion
		}
	}

	return ingestionErrors
}

// --- Helper functions for file type identification and extraction ---

type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeXML
	FileTypeZIP
	FileTypeGZ
	FileTypeTGZ
)

// identifyFileType determines the file type based on header and filename.
func identifyFileType(header []byte, filename string) FileType {
	// Check for ZIP magic number
	if len(header) >= 4 && header[0] == 0x50 && header[1] == 0x4B && header[2] == 0x03 && header[3] == 0x04 {
		return FileTypeZIP
	}
	// Check for GZIP magic number
	if len(header) >= 2 && header[0] == 0x1F && header[1] == 0x8B {
		// Check if it's a .tgz (tar.gz)
		if strings.HasSuffix(strings.ToLower(filename), ".tgz") || strings.HasSuffix(strings.ToLower(filename), ".tar.gz") {
			return FileTypeTGZ
		}
		return FileTypeGZ
	}
	// Check for XML declaration (<?xml)
	if len(header) >= 5 && strings.HasPrefix(string(header), "<?xml") {
		return FileTypeXML
	}
	// Check for UTF-8 BOM followed by XML declaration
	if len(header) >= 8 && header[0] == 0xEF && header[1] == 0xBB && header[2] == 0xBF && strings.HasPrefix(string(header[3:]), "<?xml") {
		return FileTypeXML
	}

	// Fallback to extension if header is inconclusive but extension is clear
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xml":
		return FileTypeXML
	case ".zip":
		return FileTypeZIP
	case ".gz":
		return FileTypeGZ
	case ".tgz", ".tar.gz":
		return FileTypeTGZ
	}

	return FileTypeUnknown
}

// extractXMLContent extracts XML content from various file types.
func (rp *ReportProcessor) extractXMLContent(reader io.Reader, fileType FileType) ([][]byte, error) {
	var xmlContents [][]byte

	switch fileType {
	case FileTypeXML:
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read XML file: %w", err)
		}
		xmlContents = append(xmlContents, content)
	case FileTypeGZ:
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		content, err := io.ReadAll(gzReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read gzipped content: %w", err)
		}
		xmlContents = append(xmlContents, content)
	case FileTypeZIP:
		// ZIP files need to be read from a seeker, so we need to read into a buffer first
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, reader); err != nil {
			return nil, fmt.Errorf("failed to copy zip content to buffer: %w", err)
		}
		zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			return nil, fmt.Errorf("failed to create zip reader: %w", err)
		}

		for _, f := range zipReader.File {
			if strings.HasSuffix(strings.ToLower(f.Name), ".xml") {
				rc, err := f.Open()
				if err != nil {
					log.Printf("Failed to open file %s in zip: %v", f.Name, err)
					continue
				}
				content, err := io.ReadAll(rc)
				rc.Close()
				if err != nil {
					log.Printf("Failed to read content from %s in zip: %v", f.Name, err)
					continue
				}
				xmlContents = append(xmlContents, content)
			}
		}
	case FileTypeTGZ:
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader for tgz: %w", err)
		}
		defer gzReader.Close()

		tarReader := tar.NewReader(gzReader)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break // End of tar archive
			}
			if err != nil {
				return nil, fmt.Errorf("failed to read tar header: %w", err)
			}

			if header.Typeflag == tar.TypeReg && strings.HasSuffix(strings.ToLower(header.Name), ".xml") {
				content, err := io.ReadAll(tarReader)
				if err != nil {
					log.Printf("Failed to read content from %s in tgz: %v", header.Name, err)
					continue
				}
				xmlContents = append(xmlContents, content)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported file type for extraction: %v", fileType)
	}

	return xmlContents, nil
}

// peekableReader is a wrapper around io.Reader that allows peeking at the beginning of the stream.
type peekableReader struct {
	reader io.Reader
	buffer *bytes.Buffer
	peeked bool
}

func newPeekableReader(r io.Reader) *peekableReader {
	return &peekableReader{
		reader: r,
		buffer: new(bytes.Buffer),
	}
}

func (pr *peekableReader) Read(p []byte) (n int, err error) {
	if pr.buffer.Len() > 0 {
		n, err = pr.buffer.Read(p)
		if err == io.EOF { // Buffer exhausted, read from underlying reader
			return pr.reader.Read(p)
		}
		return n, err
	}
	return pr.reader.Read(p)
}

func (pr *peekableReader) Peek(n int) ([]byte, error) {
	if pr.buffer.Len() >= n {
		return pr.buffer.Bytes()[:n], nil
	}

	// Read more into buffer if needed
	needed := n - pr.buffer.Len()
	temp := make([]byte, needed)
	readN, err := pr.reader.Read(temp)
	if err != nil && err != io.EOF {
		return nil, err
	}
	pr.buffer.Write(temp[:readN])

	if pr.buffer.Len() < n {
		return pr.buffer.Bytes(), io.EOF // Not enough data to peek 'n' bytes
	}
	return pr.buffer.Bytes()[:n], nil
}
