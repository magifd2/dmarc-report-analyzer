package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/oschwald/maxminddb-golang"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run mmdb_dumper.go <mmdb_file_path> <ip_address>")
		os.Exit(1)
	}

	mmdbFilePath := os.Args[1]
	ipStr := os.Args[2]

	db, err := maxminddb.Open(mmdbFilePath)
	if err != nil {
		log.Fatalf("Could not open database %s: %v", mmdbFilePath, err)
	}
	defer db.Close()

	ip := net.ParseIP(ipStr)
	if ip == nil {
		log.Fatalf("Invalid IP address: %s", ipStr)
	}

	var record interface{} // Use interface{} to dump raw structure
	err = db.Lookup(ip, &record)
	if err != nil {
		log.Fatalf("Lookup failed for %s: %v", ipStr, err)
	}

	if record == nil {
		fmt.Printf("No record found for IP: %s\n", ipStr)
		return
	}

	fmt.Printf("MMDB Dump for IP %s from %s:\n", ipStr, mmdbFilePath)
	fmt.Printf("%+v\n", record) // Dump the raw structure
}