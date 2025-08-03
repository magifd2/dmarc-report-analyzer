package ip_geo

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"dmarc-report-analyzer/backend/src/db"
	"dmarc-report-analyzer/backend/src/util"
)

const (
	cityDBName = "ipinfo-city.mmdb"
	asnDBName  = "ipinfo-asn.mmdb"
)

// Resolver provides methods to resolve IP addresses to geographical and ASN information.
type Resolver struct {
	CityDB *maxminddb.Reader // Exported
	AsnDB  *maxminddb.Reader // Exported
	dataDir string // Absolute path to the ip_geo data directory
	mu     sync.RWMutex // Protects access to CityDB and AsnDB
}

// NewResolver creates and initializes a new IP Geo Resolver.
// It attempts to load the MMDB files from the specified data directory.
func NewResolver(dataDir string) (*Resolver, error) { // Added dataDir parameter
	r := &Resolver{dataDir: dataDir} // Initialize dataDir
	err := r.LoadDatabases()
	if err != nil {
		log.Printf("Warning: Failed to load IP Geo databases: %v. IP resolution may be limited.", err)
		// Do not return error, allow app to start without IP geo if files are missing/corrupt
	}
	return r, nil
}

// LoadDatabases attempts to load the IP Geo MMDB files.
func (r *Resolver) LoadDatabases() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Close existing readers if any
	if r.CityDB != nil {
		r.CityDB.Close()
		r.CityDB = nil
	}
	if r.AsnDB != nil {
		r.AsnDB.Close()
		r.AsnDB = nil
	}

	// Ensure the data directory exists
	// Use r.dataDir which is already the absolute path to ip_geo directory
	if err := os.MkdirAll(r.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create IP geo database directory %s: %w", r.dataDir, err)
	}

	cityDBPath := filepath.Join(r.dataDir, cityDBName) // Use r.dataDir
	asnDBPath := filepath.Join(r.dataDir, asnDBName)   // Use r.dataDir

	var errs []error

	cityDB, err := maxminddb.Open(cityDBPath)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to open city database %s: %w", cityDBPath, err))
	} else {
		r.CityDB = cityDB
	}

	asnDB, err := maxminddb.Open(asnDBPath)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to open ASN database %s: %w", asnDBPath, err))
	} else {
		r.AsnDB = asnDB
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors loading IP Geo databases: %v", errs)
	}
	log.Println("Successfully loaded IP Geo databases.")
	return nil
}

// ResolveIP resolves an IP address to GeoIP, ASN, and DNS information.
func (r *Resolver) ResolveIP(ipStr string) (*db.IPInfo, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address format: %s", ipStr)
	}

	ipInfo := &db.IPInfo{
		IPAddress: ipStr,
		LastUpdated: time.Now().Unix(),
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Resolve City/Country
	if r.CityDB != nil {
		var cityRecord struct {
			Country struct {
				ISOCode string `maxminddb:"iso_code"`
				Names   map[string]string `maxminddb:"names"`
			} `maxminddb:"country"`
			City struct {
				Names map[string]string `maxminddb:"names"`
			} `maxminddb:"city"`
		}
		err := r.CityDB.Lookup(ip, &cityRecord)
		if err != nil {
			log.Printf("City lookup failed for %s: %v", ipStr, err)
		} else {
			ipInfo.CountryCode = cityRecord.Country.ISOCode
			ipInfo.CountryName = cityRecord.Country.Names["en"] // Assuming English name
			ipInfo.CityName = cityRecord.City.Names["en"]
		}
	}

	// Resolve ASN
	if r.AsnDB != nil {
		var asnRecord struct {
			ASN        uint `maxminddb:"asn"`
			Organization string `maxminddb:"organization"`
		}
		err := r.AsnDB.Lookup(ip, &asnRecord)
		if err != nil {
			log.Printf("ASN lookup failed for %s: %v", ipStr, err)
		} else {
			ipInfo.ASNNumber = int(asnRecord.ASN)
			ipInfo.ASNOrganization = asnRecord.Organization
		}
	}

	// Resolve Hostname (PTR), ReversedHostname, and ApexDomain
	hostname := util.ResolvePTR(ipStr)
	ipInfo.Hostname = hostname
	if hostname != "N/A" && hostname != "" {
		ipInfo.ReversedHostname = util.ReverseDomain(hostname)
		ipInfo.ApexDomain = util.GetApexDomain(hostname)
	} else {
		ipInfo.ReversedHostname = "N/A"
		ipInfo.ApexDomain = "N/A"
	}

	return ipInfo, nil
}

// ImportMMDBFile copies a given MMDB file to the ip_geo data directory.
// It now accepts the absolute path to the ip_geo data directory.
func ImportMMDBFile(srcPath string, destDataDir string) error { // Added destDataDir parameter
	if err := os.MkdirAll(destDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDataDir, err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source MMDB file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	// Determine destination filename based on source filename
	destFileName := filepath.Base(srcPath)
	if strings.Contains(strings.ToLower(destFileName), "city") {
		destFileName = cityDBName
	} else if strings.Contains(strings.ToLower(destFileName), "asn") {
		destFileName = asnDBName
	} else {
		return fmt.Errorf("unrecognized MMDB file type for %s. Expected 'city' or 'asn' in filename.", srcPath)
	}

	destPath := filepath.Join(destDataDir, destFileName) // Use destDataDir
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination MMDB file %s: %w", destPath, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy MMDB file from %s to %s: %w", srcPath, destPath, err)
	}

	log.Printf("Successfully imported MMDB file from %s to %s", srcPath, destPath)
	return nil
}