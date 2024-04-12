package types_test

import (
	"strings"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestValidateCIDR(t *testing.T) {
	for _, tc := range []struct {
		cidr      string
		expectErr bool
	}{
		{cidr: "10.1.0.0/16"},
		{cidr: "2001:0db8::/32"},
		{cidr: "10.1.0.0/16,2001:0db8::/32"},
		{cidr: "", expectErr: true},
		{cidr: "bananas", expectErr: true},
		{cidr: "fd01::/64,fd02::/64,fd03::/64", expectErr: true},
	} {
		t.Run(tc.cidr, func(t *testing.T) {
			t.Run("Pod", func(t *testing.T) {
				g := NewWithT(t)
				config := types.ClusterConfig{
					Network: types.Network{
						PodCIDR:     vals.Pointer(tc.cidr),
						ServiceCIDR: vals.Pointer("10.1.0.0/16"),
					},
				}
				err := config.Validate()
				if tc.expectErr {
					g.Expect(err).To(HaveOccurred())
				} else {
					g.Expect(err).To(BeNil())
				}
			})
			t.Run("Service", func(t *testing.T) {
				g := NewWithT(t)
				config := types.ClusterConfig{
					Network: types.Network{
						PodCIDR:     vals.Pointer("10.1.0.0/16"),
						ServiceCIDR: vals.Pointer(tc.cidr),
					},
				}
				err := config.Validate()
				if tc.expectErr {
					g.Expect(err).To(HaveOccurred())
				} else {
					g.Expect(err).To(BeNil())
				}
			})
		})
	}
}

func TestValidateLoadbalancer(t *testing.T) {
	for _, tc := range []struct {
		cidr      []string
		expectErr bool
	}{
		{cidr: []string{"10.3.0.0/16"}},
		{cidr: []string{"2001:0db8::/32"}},
		{cidr: []string{"10.3.0.0/16", "2001:0db8::/32"}},
		{cidr: []string{"10.3.0.10-10.3.0.20"}},
		{cidr: []string{"10.3.0.32/28", "10.3.0.10-10.3.0.20"}},
		{cidr: []string{"2001:0db8::0-2001:0db8::10"}},
		{cidr: []string{"10.3.0.32/28", "2001:0db8::0-2001:0db8::10"}},
		{cidr: []string{""}, expectErr: true},
		{cidr: []string{"bananas"}, expectErr: true},
		{cidr: []string{"fd01::/64,fd02::/64,fd03::/64"}, expectErr: true},
		{cidr: []string{"10.3.0.10-10.3.0.300"}, expectErr: true},
		{cidr: []string{"10.3.0.10-10.3.0.7"}, expectErr: true},
		{cidr: []string{"2001:0db8::0-2001:0db8::gg"}, expectErr: true},
		{cidr: []string{"", "10.3.0.10-10.3.0.300"}, expectErr: true},
		{cidr: []string{"10.3.0.10-10.3.0.12-10.3.0.15"}, expectErr: true},
	} {
		t.Run(strings.Join(tc.cidr, ","), func(t *testing.T) {
			t.Run("LoadBalancer", func(t *testing.T) {
				g := NewWithT(t)
				config := types.ClusterConfig{
					Network: types.Network{
						Enabled:     vals.Pointer(true),
						PodCIDR:     vals.Pointer("10.2.0.0/16"),
						ServiceCIDR: vals.Pointer("10.1.0.0/16"),
					},
					LoadBalancer: types.LoadBalancer{
						Enabled: vals.Pointer(true),
						CIDRs:   vals.Pointer(tc.cidr),
					},
				}
				err := config.Validate()
				if tc.expectErr {
					g.Expect(err).Should(HaveOccurred())
				} else {
					g.Expect(err).ShouldNot(HaveOccurred())
				}
			})
		})
	}
}
