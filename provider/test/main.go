package main

import (
	"fmt"
	"strings"
)

func main() {
	domain := "sub.sub.sub.domain.com"

	// Split the domain into parts
	parts := strings.Split(domain, ".")

	// Prepare to collect all parent domains
	var parentDomains []string

	// Construct each parent domain starting from the full domain
	for i := range parts {
		// Join parts from i to end
		parentDomain := strings.Join(parts[i:], ".")
		parentDomains = append(parentDomains, parentDomain)
	}

	// Display all parent domains
	for _, parentDomain := range parentDomains {
		fmt.Println(parentDomain)
	}
}
