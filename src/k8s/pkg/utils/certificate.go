package utils

import (
	"net"
	"strings"
)

func GetExtraSANsFromString(extraSANs string) []string {
	// TODO: Add validation for the extraSANs
	return strings.Split(extraSANs, ",")
}

func SeparateSANs(extraSANs []string) ([]net.IP, []string) {
	var ipSANs []net.IP
	var dnsSANs []string

	for _, san := range extraSANs {
		if ip := net.ParseIP(san); ip != nil {
			ipSANs = append(ipSANs, ip)
		} else {
			dnsSANs = append(dnsSANs, san)
		}
	}

	return ipSANs, dnsSANs
}
