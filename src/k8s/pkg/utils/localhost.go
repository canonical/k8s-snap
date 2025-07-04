package utils

import (
	"fmt"
	"net"
)

// GetLocalhostAddress returns the loopback address (IPv4 or IPv6) of the local machine.
// It checks the loopback interface "lo" and returns "127.0.0.1" if available,
// or "::1" if only IPv6 is available. If neither is found, it returns an error.
func GetLocalhostAddress() (net.IP, error) {
	iface, err := net.InterfaceByName("lo")
	if err != nil {
		return nil, fmt.Errorf("failed to get loopback interface: %w", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses for loopback interface: %w", err)
	}

	var ipv6 net.IP
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		if ip.Equal(net.ParseIP("127.0.0.1")) {
			return ip, nil
		}
		if ip.Equal(net.ParseIP("::1")) {
			ipv6 = ip
		}
	}

	if ipv6 != nil {
		return ipv6, nil
	}

	return nil, fmt.Errorf("no loopback address found")
}
