package utils

import (
	"fmt"
	"net"

	"github.com/canonical/lxd/lxd/util"
)

// GetDefaultIP returns the IP address of the default interface.
func GetDefaultIP() (net.IP, error) {
	parsed := net.ParseIP(util.NetworkInterfaceAddress())
	if parsed == nil {
		return nil, fmt.Errorf("failed to get the IP address of the default interface")
	}

	return parsed, nil
}
