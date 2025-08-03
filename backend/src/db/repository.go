package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Repository provides methods for interacting with the database.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Repository instance.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveReport saves a DMARC report to the database.
func (r *Repository) SaveReport(report *Report) (int64, error) {
	stmt, err := r.db.Prepare(`
		INSERT INTO reports (xml_hash, original_xml, org_name, report_id, date_range_begin, date_range_end, domain, adkim, aspf, p, sp, pct)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement for saving report: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		report.XMLHash, report.OriginalXML, report.OrgName, report.ReportID,
		report.DateRangeBegin, report.DateRangeEnd, report.Domain, report.ADKIM,
		report.ASPF, report.P, report.SP, report.PCT,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement for saving report: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID for report: %w", err)
	}
	report.ID = id
	return id, nil
}

// ReportExistsByHash checks if a report with the given XML hash already exists.
func (r *Repository) ReportExistsByHash(xmlHash string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM reports WHERE xml_hash = ?", xmlHash).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to query report by hash: %w", err)
	}
	return count > 0, nil
}

// GetReports retrieves a list of DMARC reports with pagination and sorting.
func (r *Repository) GetReports(limit, offset int, sortBy, sortOrder string) ([]Report, int, error) {
	var reports []Report
	var totalCount int

	// Get total count
	err := r.db.QueryRow("SELECT COUNT(*) FROM reports").Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total report count: %w", err)
	}

	// Validate sortBy and sortOrder to prevent SQL injection
	validSortBy := map[string]bool{
		"id": true, "org_name": true, "report_id": true, "date_range_begin": true, "domain": true,
	}
	validSortOrder := map[string]bool{
		"asc": true, "desc": true,
	}

	if !validSortBy[sortBy] {
		sortBy = "date_range_begin" // Default sort by
	}
	if !validSortOrder[strings.ToLower(sortOrder)] {
		sortOrder = "desc" // Default sort order
	}

	query := fmt.Sprintf(`
		SELECT id, xml_hash, original_xml, org_name, report_id, date_range_begin, date_range_end, domain, adkim, aspf, p, sp, pct
		FROM reports
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, sortBy, sortOrder)

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var report Report
		err := rows.Scan(
			&report.ID, &report.XMLHash, &report.OriginalXML, &report.OrgName, &report.ReportID,
			&report.DateRangeBegin, &report.DateRangeEnd, &report.Domain, &report.ADKIM,
			&report.ASPF, &report.P, &report.SP, &report.PCT,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan report row: %w", err)
		}
		reports = append(reports, report)
	}

	return reports, totalCount, nil
}

// GetReportByID retrieves a single DMARC report by its ID.
func (r *Repository) GetReportByID(id int64) (*Report, error) {
	var report Report
	err := r.db.QueryRow(`
		SELECT id, xml_hash, original_xml, org_name, report_id, date_range_begin, date_range_end, domain, adkim, aspf, p, sp, pct
		FROM reports
		WHERE id = ?
	`, id).Scan(
		&report.ID, &report.XMLHash, &report.OriginalXML, &report.OrgName, &report.ReportID,
		&report.DateRangeBegin, &report.DateRangeEnd, &report.Domain, &report.ADKIM,
		&report.ASPF, &report.P, &report.SP, &report.PCT,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Report not found
		}
		return nil, fmt.Errorf("failed to query report by ID: %w", err)
	}
	return &report, nil
}

// GetUserByUsername retrieves a user by their username.
func (r *Repository) GetUserByUsername(username string) (*User, error) {
	var user User
	err := r.db.QueryRow("SELECT id, username, password_hash, created_at FROM users WHERE username = ?", username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to query user by username: %w", err)
	}
	return &user, nil
}

// UpdateUser updates a user's information in the database.
func (r *Repository) UpdateUser(user *User) error {
	stmt, err := r.db.Prepare(`
		UPDATE users
		SET username = ?, password_hash = ?
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for updating user: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Username, user.PasswordHash, user.ID)
	if err != nil {
		return fmt.Errorf("failed to execute statement for updating user: %w", err)
	}
	return nil
}

// SaveRecord saves a DMARC record to the database.
func (r *Repository) SaveRecord(record *Record) error {
	stmt, err := r.db.Prepare(`
		INSERT INTO records (report_id, source_ip, count, header_from, disposition, dkim_result, spf_result)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for saving record: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		record.ReportID, record.SourceIP, record.Count, record.HeaderFrom,
		record.Disposition, record.DKIMResult, record.SPFResult,
	)
	if err != nil {
		return fmt.Errorf("failed to execute statement for saving record: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID for record: %w", err)
	}
	record.ID = id
	return nil
}

// SaveOrUpdateIPInfo saves or updates IP information in the database.
func (r *Repository) SaveOrUpdateIPInfo(info *IPInfo) error {
	stmt, err := r.db.Prepare(`
		INSERT INTO ip_info (ip_address, country_code, country_name, city_name, asn_number, asn_organization, hostname, reversed_hostname, apex_domain, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ip_address) DO UPDATE SET
			country_code = EXCLUDED.country_code,
			country_name = EXCLUDED.country_name,
			city_name = EXCLUDED.city_name,
			asn_number = EXCLUDED.asn_number,
			asn_organization = EXCLUDED.asn_organization,
			hostname = EXCLUDED.hostname,
			reversed_hostname = EXCLUDED.reversed_hostname,
			apex_domain = EXCLUDED.apex_domain,
			last_updated = EXCLUDED.last_updated
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for saving/updating IP info: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		info.IPAddress, info.CountryCode, info.CountryName, info.CityName,
		info.ASNNumber, info.ASNOrganization, info.Hostname, info.ReversedHostname,
		info.ApexDomain, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to execute statement for saving/updating IP info: %w", err)
	}
	return nil
}

// SaveIngestionError saves an ingestion error to the database.
func (r *Repository) SaveIngestionError(errInfo *IngestionError) error {
	stmt, err := r.db.Prepare(`
		INSERT INTO ingestion_errors (filename, xml_hash, error_type, message, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for saving ingestion error: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		errInfo.Filename, errInfo.XMLHash, errInfo.ErrorType, errInfo.Message, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to execute statement for saving ingestion error: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID for ingestion error: %w", err)
	}
	errInfo.ID = id
	return nil
}