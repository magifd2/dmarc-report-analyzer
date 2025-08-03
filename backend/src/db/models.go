package db

// Report represents a DMARC aggregate report.
type Report struct {
	ID             int64  `db:"id"`
	XMLHash        string `db:"xml_hash"`
	OriginalXML    string `db:"original_xml"`
	OrgName        string `db:"org_name"`
	ReportID       string `db:"report_id"`
	DateRangeBegin int64  `db:"date_range_begin"`
	DateRangeEnd   int64  `db:"date_range_end"`
	Domain         string `db:"domain"`
	ADKIM          string `db:"adkim"`
	ASPF           string `db:"aspf"`
	P              string `db:"p"`
	SP             string `db:"sp"`
	PCT            int    `db:"pct"`
}

// Record represents a single record within a DMARC report.
type Record struct {
	ID          int64  `db:"id"`
	ReportID    int64  `db:"report_id"`
	SourceIP    string `db:"source_ip"`
	Count       int    `db:"count"`
	HeaderFrom  string `db:"header_from"`
	Disposition string `db:"disposition"`
	DKIMResult  string `db:"dkim_result"`
	SPFResult   string `db:"spf_result"`
}

// IPInfo represents geographical and ASN information for an IP address.
type IPInfo struct {
	IPAddress       string `db:"ip_address"`
	CountryCode     string `db:"country_code"`
	CountryName     string `db:"country_name"`
	CityName        string `db:"city_name"`
	ASNNumber       int    `db:"asn_number"`
	ASNOrganization string `db:"asn_organization"`
	Hostname        string `db:"hostname"`
	ReversedHostname string `db:"reversed_hostname"`
	ApexDomain      string `db:"apex_domain"`
	LastUpdated     int64  `db:"last_updated"`
}

// IngestionError represents an error that occurred during DMARC report ingestion.
type IngestionError struct {
	ID        int64  `db:"id"`
	Filename  string `db:"filename"`
	XMLHash   string `db:"xml_hash"`
	ErrorType string `db:"error_type"`
	Message   string `db:"message"`
	Timestamp int64  `db:"timestamp"`
}

// User represents a user of the application.
type User struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	CreatedAt    int64  `db:"created_at"`
}

// Setting represents a key-value pair for application settings.
type Setting struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}