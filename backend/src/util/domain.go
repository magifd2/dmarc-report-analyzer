package util

import (
	"log"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// ReverseDomain reverses the components of a domain name.
// e.g., "www.example.com" -> "com.example.www"
func ReverseDomain(domain string) string {
	parts := strings.Split(domain, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".")
}

// GetApexDomain extracts the apex domain (eTLD+1) from a given domain name.
// e.g., "www.example.com" -> "example.com"
// e.g., "mail.sub.example.co.jp" -> "example.co.jp"
func GetApexDomain(domain string) string {
	apex, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		log.Printf("Error getting apex domain for %s: %v. Returning original domain.", domain, err)
		return domain // Fallback to original domain on error
	}
	return apex
}
