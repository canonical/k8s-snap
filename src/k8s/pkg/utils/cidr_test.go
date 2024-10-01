package utils_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/util"
	. "github.com/onsi/gomega"
)

func TestGetFirstIP(t *testing.T) {
	for _, tc := range []struct {
		cidr string
		ip   string
	}{
		{cidr: "10.152.183.0/24", ip: "10.152.183.1"},
		{cidr: "10.152.183.10/24", ip: "10.152.183.1"},
		{cidr: "10.100.0.0/16", ip: "10.100.0.1"},
		{cidr: "fd01::/64", ip: "fd01::1"},
		// TODO: do we need more test cases?
	} {
		t.Run(tc.cidr, func(t *testing.T) {
			g := NewWithT(t)
			ip, err := utils.GetFirstIP(tc.cidr)
			g.Expect(err).To(BeNil())
			g.Expect(ip.String()).To(Equal(tc.ip))
		})
	}
}

func TestGetKubernetesServiceIPsFromServiceCIDRs(t *testing.T) {
	// Test valid subnet cidr strings
	t.Run("ValidCIDR", func(t *testing.T) {
		for _, tc := range []struct {
			cidr string
			ips  []string
		}{
			{cidr: "10.152.183.0/24", ips: []string{"10.152.183.1"}},
			{cidr: "fd01::/64", ips: []string{"fd01::1"}},
			{cidr: "10.152.183.0/24,fd01::/64", ips: []string{"10.152.183.1", "fd01::1"}},
		} {
			t.Run(tc.cidr, func(t *testing.T) {
				g := NewWithT(t)
				i, err := utils.GetKubernetesServiceIPsFromServiceCIDRs(tc.cidr)
				ips := make([]string, len(i))
				for idx, v := range i {
					ips[idx] = v.String()
				}

				g.Expect(err).To(BeNil())
				g.Expect(ips).To(Equal(tc.ips))
			})
		}
	})

	t.Run("InvalidCIDR", func(t *testing.T) {
		for _, tc := range []struct {
			cidr string
		}{
			{cidr: "fd01::/64,fd02::/64,fd03::/64"},
			{cidr: "bananas"},
		} {
			t.Run(tc.cidr, func(t *testing.T) {
				g := NewWithT(t)
				_, err := utils.GetKubernetesServiceIPsFromServiceCIDRs(tc.cidr)

				g.Expect(err).ToNot(BeNil())
			})
		}
	})
}

func TestParseAddressString(t *testing.T) {
	g := NewWithT(t)

	// Seed the default address
	defaultAddress := util.NetworkInterfaceAddress()
	ip := net.ParseIP(defaultAddress)
	subnetMask := net.CIDRMask(24, 32)
	networkAddress := ip.Mask(subnetMask)
	// Infer the CIDR notation
	networkAddressCIDR := fmt.Sprintf("%s/24", networkAddress.String())

	for _, tc := range []struct {
		name    string
		address string
		port    int64
		want    string
		wantErr bool
	}{
		{name: "EmptyAddress", address: "", port: 8080, want: fmt.Sprintf("%s:8080", defaultAddress), wantErr: false},
		{name: "CIDR", address: networkAddressCIDR, port: 8080, want: fmt.Sprintf("%s:8080", defaultAddress), wantErr: false},
		{name: "CIDRAndPort", address: fmt.Sprintf("%s:9090", networkAddressCIDR), port: 8080, want: fmt.Sprintf("%s:9090", defaultAddress), wantErr: false},
		{name: "IPv4", address: "10.0.0.10", port: 8080, want: "10.0.0.10:8080", wantErr: false},
		{name: "IPv4AndPort", address: "10.0.0.10:9090", port: 8080, want: "10.0.0.10:9090", wantErr: false},
		{name: "NonMatchingCIDR", address: "10.10.5.0/24", port: 8080, want: "", wantErr: true},
		{name: "IPv6", address: "fe80::1:234", port: 8080, want: "[fe80::1:234]:8080", wantErr: false},
		{name: "IPv6AndPort", address: "[fe80::1:234]:9090", port: 8080, want: "[fe80::1:234]:9090", wantErr: false},
		{name: "InvalidPort", address: "127.0.0.1:invalid-port", port: 0, want: "", wantErr: true},
		{name: "PortOutOfBounds", address: "10.0.0.10:70799", port: 8080, want: "", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := utils.ParseAddressString(tc.address, tc.port)
			if tc.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(got).To(Equal(tc.want))
			}
		})
	}
}

func TestParseCIDRs(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		input        string
		expectedIPv4 string
		expectedIPv6 string
		expectedErr  bool
	}{
		{
			input:        "192.168.1.0/24",
			expectedIPv4: "192.168.1.0/24",
			expectedIPv6: "",
		},
		{
			input:        "2001:db8::/32",
			expectedIPv4: "",
			expectedIPv6: "2001:db8::/32",
		},
		{
			input:        "192.168.1.0/24,2001:db8::/32",
			expectedIPv4: "192.168.1.0/24",
			expectedIPv6: "2001:db8::/32",
		},
		{
			input:        "192.168.1.0/24,invalidCIDR",
			expectedIPv4: "",
			expectedIPv6: "",
			expectedErr:  true,
		},
		{
			input:        "192.168.1.0/24,2001:db8::/32,10.0.0.0/8",
			expectedIPv4: "",
			expectedIPv6: "",
			expectedErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			ipv4CIDR, ipv6CIDR, err := utils.SplitCIDRStrings(tc.input)
			if tc.expectedErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).To(BeNil())
				Expect(ipv4CIDR).To(Equal(tc.expectedIPv4))
				Expect(ipv6CIDR).To(Equal(tc.expectedIPv6))
			}
		})
	}
}

func TestIsIPv4(t *testing.T) {
	tests := []struct {
		address  string
		expected bool
	}{
		{"192.168.1.1:80", true},
		{"127.0.0.1", true},
		{"::1", false},
		{"[fe80::1]:80", false},
		{"256.256.256.256", false}, // Invalid IPv4 address
	}

	for _, tc := range tests {
		t.Run(tc.address, func(t *testing.T) {
			g := NewWithT(t)
			result := utils.IsIPv4(tc.address)
			g.Expect(result).To(Equal(tc.expected))
		})
	}
}

func TestToIPString(t *testing.T) {
	tests := []struct {
		ip       net.IP
		expected string
	}{
		{net.ParseIP("192.168.1.1"), "192.168.1.1"},
		{net.ParseIP("::1"), "[::1]"},
		{net.ParseIP("fe80::1"), "[fe80::1]"},
		{net.ParseIP("127.0.0.1"), "127.0.0.1"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			g := NewWithT(t)
			result := utils.ToIPString(tc.ip)
			g.Expect(result).To(Equal(tc.expected))
		})
	}
}
