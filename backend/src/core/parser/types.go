package parser

import (
	"encoding/xml"
	"fmt"
	"time"
	// "log" // Removed debugging log import
)

// Feedback is the root element of a DMARC aggregate report.
type Feedback struct {
	XMLName        xml.Name       `xml:"feedback"`
	ReportMetadata ReportMetadata `xml:"report_metadata"`
	PolicyPublished PolicyPublished `xml:"policy_published"`
	Records        []Record       `xml:"record"`
}

// ReportMetadata contains metadata about the report.
type ReportMetadata struct {
	OrgName   string `xml:"org_name"`
	Email     string `xml:"email"`
	ReportID  string `xml:"report_id"`
	DateRange struct {
		Begin int64 `xml:"begin"`
		End   int64 `xml:"end"`
	} `xml:"date_range"`
	// Other optional fields like error, extra_contact_info
}

// PolicyPublished contains the DMARC policy for the domain.
type PolicyPublished struct {
	Domain string `xml:"domain"`
	ADKIM  string `xml:"adkim"`
	ASPF   string `xml:"aspf"`
	P      string `xml:"p"`
	SP     string `xml:"sp"`
	PCT    int    `xml:"pct"`
	// Other optional fields like fo
}

// PolicyEvaluated contains the DMARC policy evaluation results for a record.
type PolicyEvaluated struct {
	Disposition string `xml:"disposition"`
	DKIM        string `xml:"dkim"`
	SPF         string `xml:"spf"`
	// Other optional fields like reasons, enforcement_uri
}

// RecordRow contains information about the source IP, count, and policy evaluation.
type RecordRow struct {
	SourceIP        string          `xml:"source_ip"`
	Count           int             `xml:"count"`
	PolicyEvaluated PolicyEvaluated `xml:"policy_evaluated"` // Moved here
	// Other optional fields like reason
}

// Identifiers contains identifying information for the email.
type Identifiers struct {
	HeaderFrom string `xml:"header_from"`
}

// AuthResults contains authentication results (DKIM and SPF).
type AuthResults struct {
	DKIM []AuthResultDKIM `xml:"dkim"`
	SPF  []AuthResultSPF  `xml:"spf"`
}

// Record contains information about a specific set of email streams.
type Record struct {
	Row             RecordRow       `xml:"row"`
	Identifiers     Identifiers     `xml:"identifiers"`
	// PolicyEvaluated removed from here
	AuthResults     AuthResults     `xml:"auth_results"`
}

// AuthResultDKIM represents a DKIM authentication result.
type AuthResultDKIM struct {
	Domain   string `xml:"domain"`
	Result   string `xml:"result"`
	Selector string `xml:"selector"`
	// Other optional fields like human_result
}

// AuthResultSPF represents an SPF authentication result.
type AuthResultSPF struct {
	Domain string `xml:"domain"`
	Result string `xml:"result"`
	// Other optional fields like human_result
}

// Validate performs strict validation on the DMARC Feedback report.
func (f *Feedback) Validate() error {
	if f.ReportMetadata.OrgName == "" {
		return fmt.Errorf("report_metadata.org_name is missing")
	}
	if f.ReportMetadata.ReportID == "" {
		return fmt.Errorf("report_metadata.report_id is missing")
	}
	if f.ReportMetadata.DateRange.Begin == 0 || f.ReportMetadata.DateRange.End == 0 {
		return fmt.Errorf("report_metadata.date_range (begin or end) is missing or zero")
	}
	if f.ReportMetadata.DateRange.Begin >= f.ReportMetadata.DateRange.End {
		return fmt.Errorf("report_metadata.date_range.begin must be less than date_range.end")
	}
	// Check if timestamps are reasonable (e.g., not in the future, not too far in the past)
	if time.Unix(f.ReportMetadata.DateRange.End, 0).After(time.Now().Add(24 * time.Hour)) { // Allow some future buffer
		return fmt.Errorf("report_metadata.date_range.end is in the far future")
	}

	if f.PolicyPublished.Domain == "" {
		return fmt.Errorf("policy_published.domain is missing")
	}
	if f.PolicyPublished.P == "" {
		return fmt.Errorf("policy_published.p is missing")
	}
	if f.PolicyPublished.PCT < 0 || f.PolicyPublished.PCT > 100 {
		return fmt.Errorf("policy_published.pct must be between 0 and 100")
	}

	if len(f.Records) == 0 {
		return fmt.Errorf("no records found in the report")
	}

	for i, record := range f.Records {
		if record.Row.SourceIP == "" {
			return fmt.Errorf("record[%d].row.source_ip is missing", i)
		}
		if record.Row.Count <= 0 {
			return fmt.Errorf("record[%d].row.count must be positive", i)
		}
		if record.Identifiers.HeaderFrom == "" {
			return fmt.Errorf("record[%d].identifiers.header_from is missing", i)
		}
		// Access through record.Row
		if record.Row.PolicyEvaluated.Disposition == "" {
			return fmt.Errorf("record[%d].row.policy_evaluated.disposition is missing", i)
		}
		// Add more specific validations for disposition, DKIM/SPF results, etc.
	}

	return nil
}