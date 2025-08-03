package util

import (
	"context"
	"log"
	"net"
	"time"
)

// ResolvePTR performs a reverse DNS lookup (PTR record) for the given IP address.
// It returns the resolved hostname or "N/A" if the lookup fails or no PTR record is found.
// A timeout is applied to prevent long-running DNS queries.
func ResolvePTR(ipAddr string) string {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		log.Printf("DNS: Invalid IP address format: %s", ipAddr)
		return "N/A"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 5 seconds timeout
	defer cancel()

	names, err := net.DefaultResolver.LookupAddr(ctx, ipAddr)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok {
			if dnsErr.IsNotFound {
				log.Printf("DNS: No PTR record found for %s", ipAddr)
				return "N/A"
			}
			if dnsErr.IsTimeout {
				log.Printf("DNS: Lookup timed out for %s", ipAddr)
				return "N/A"
			}
			log.Printf("DNS: DNSError for %s: %v", ipAddr, dnsErr)
		} else {
			log.Printf("DNS: Unexpected error for %s: %v", ipAddr, err)
		}
		return "N/A"
	}

	if len(names) > 0 {
		return names[0]
	}

	return "N/A"
}
