package utils

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetLocalhostAddress_ReturnsIPv4OrIPv6(t *testing.T) {
	g := NewWithT(t)

	ip, err := GetLocalhostAddress()
	g.Expect(err).ToNot(HaveOccurred(), "expected no error when getting localhost address, got %v", err)
	g.Expect(ip).NotTo(BeNil(), "expected a non-nil IP address")

	// Should be either 127.0.0.1 or ::1
	isIPv4 := ip.Equal(net.ParseIP("127.0.0.1"))
	isIPv6 := ip.Equal(net.ParseIP("::1"))
	g.Expect(isIPv4 || isIPv6).To(BeTrue(), "expected IP to be 127.0.0.1 or ::1, got %s", ip.String())
}
