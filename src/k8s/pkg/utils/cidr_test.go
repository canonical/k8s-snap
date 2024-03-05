package utils_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/utils"
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

	// Test invalid cidrs
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
}
