package utils

import (
	"fmt"
	"math/big"
	"net"
)

// GetFirstIP returns the first IP address of a subnet. Use big.Int so that it can handle both IPv4 and IPv6 addreses.
func GetFirstIP(subnet string) (net.IP, error) {
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("%q is not a valid subnet CIDR: %w", subnet, err)
	}
	r := big.NewInt(0).Add(
		big.NewInt(0).SetBytes(cidr.IP.To16()),
		big.NewInt(1),
	).Bytes()
	r = append(make([]byte, 16), r...)
	return net.IP(r[len(r)-16:]), nil
}
