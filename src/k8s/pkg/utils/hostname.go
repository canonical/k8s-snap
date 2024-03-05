package utils

import (
	"fmt"
	"strings"

	"golang.org/x/net/idna"
)

// CleanHostname sanitises hostnames.
func CleanHostname(hostname string) (string, error) {
	clean, err := idna.Lookup.ToASCII(hostname)
	if err != nil {
		return "", fmt.Errorf("failed to parse hostname %q: %w", hostname, err)
	}
	if strings.HasSuffix(hostname, ".") {
		return "", fmt.Errorf("hostname cannot end with a dot (%q)", ".")
	}
	return clean, nil
}
