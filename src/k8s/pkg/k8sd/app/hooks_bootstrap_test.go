package app

import "testing"

func TestValidateCIDROverlap(t *testing.T) {
	tests := []struct {
		name        string
		podCIDR     string
		serviceCIDR string
		expectErr   bool
	}{
		{"Empty", "", "", true},
		{"EmptyServiceCIDR", "192.168.1.0/24", "", true},
		{"EmptyPodCIDR", "", "192.168.1.0/24", true},
		//IPv4
		{"SameIPv4CIDRs", "192.168.100.0/24", "192.168.100.0/24", true},
		{"OverlappingIPv4CIDRs", "10.2.0.13/24", "10.2.0.0/24", true},
		{"IPv4CIDRMinimumSize", "192.168.1.1/32", "10.0.0.0/24", false},
		{"InvalidIPv4CIDRFormat", "192.168.1.1/33", "10.0.0.0/24", true},
		{"MaxSizeIPv4CIDRs", "0.0.0.0/0", "0.0.0.0/0", true},
		//IPv6
		{"SameIPv6CIDRs", "fe80::1/32", "fe80::1/32", true},
		{"OverlappingIPv6CIDRs", "fe80::/48", "fe80::dead/48", true},
		{"IPv6CIDRMinimumSize", "2001:db8::1/128", "fc00::/32", false},
		{"InvalidIPv6CIDRFormat", "2001:db8::1/129", "fc00::/64", true},
		{"MaxSizeIPv6CIDRs", "::/0", "::/0", true},
		//Mixed
		{"IPv4AndIPv6MixedCIDRs", "192.168.0.0/16", "2001:db8::/32", false},
		{"OnlyInvalidIPv6CIDR", "", "2001:db8::/65", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateCIDROverlap(tc.podCIDR, tc.serviceCIDR); (err != nil) != tc.expectErr {
				t.Errorf("validateCIDROverlap() error = %v, expectErr %v", err, tc.expectErr)
			}
		})
	}
}

func TestValidateCIDRSize(t *testing.T) {
	tests := []struct {
		name      string
		cidr      string
		expectErr bool
	}{
		{"Empty", "", true},
		{"DualstackCIDRValid", "192.168.2.0/24,fe80::/128", false},
		{"DualstackCIDRPrefixAtLimit", "192.168.2.0/24,fe80::/108", true},
		{"IPv6CIDRPrefixBiggerThanLimit", "fe80::/64", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateIPv6CIDRSize(tc.cidr); (err != nil) != tc.expectErr {
				t.Errorf("validateCIDRSize() error = %v, expectErr %v", err, tc.expectErr)
			}
		})
	}
}
