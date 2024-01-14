package utils

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

var defaultResolvConfs = []string{
	"/etc/resolv.conf",
	"/run/systemd/resolve/resolv.conf",
}

var nameserverRegex = regexp.MustCompile(`(?m)^\s*nameserver\s+(\S*)\s*$`)

// isValidNonLoopbackIP checks if the provided IP address is valid and not a loopback address.
func isValidNonLoopbackIP(address string) bool {
	// IPv6 addresses may contain zone, e.g. "::1%2". Drop the '%' suffix, if any.
	splitToDropScopeIfAny := strings.SplitN(address, "%", 2)
	cleanIP := splitToDropScopeIfAny[0]

	ip := net.ParseIP(cleanIP)
	return ip != nil && !ip.IsLoopback()
}

// LocateValidResolvConf searches the provided resolv.conf files for one containing only non-loopback addresses.
func LocateValidResolvConf() (string, error) {
	for _, path := range defaultResolvConfs {
		isValid := processResolvConf(path)
		if isValid {
			return path, nil
		}
	}
	return "", fmt.Errorf("no suitable resolv.conf file found")
}

// processResolvConf processes a single resolv.conf file and checks if it contains only non-loopback addresses.
func processResolvConf(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	matches := nameserverRegex.FindAllStringSubmatch(string(content), -1)
	if len(matches) == 0 {
		return false
	}

	for _, match := range matches {
		if len(match) != 2 {
			return false
		}
		if !isValidNonLoopbackIP(match[1]) {
			return false
		}
	}

	return true
}
