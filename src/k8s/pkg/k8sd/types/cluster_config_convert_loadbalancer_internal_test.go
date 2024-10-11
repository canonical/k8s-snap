package types

import (
	"testing"

	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

var loadBalancerCIDRTestCases = []struct {
	name           string
	apiCIDRs       []string
	internalCIDRs  []string
	internalRanges []LoadBalancer_IPRange
	expectErr      bool
}{
	{
		name:           "IPv4CIDR",
		apiCIDRs:       []string{"10.3.0.0/16"},
		internalCIDRs:  []string{"10.3.0.0/16"},
		internalRanges: []LoadBalancer_IPRange{},
	},
	{
		name:           "IPv6CIDR",
		apiCIDRs:       []string{"2001:0db8::/32"},
		internalCIDRs:  []string{"2001:0db8::/32"},
		internalRanges: []LoadBalancer_IPRange{},
	},
	{
		name:           "IPv4AndIPv6CIDRs",
		apiCIDRs:       []string{"10.3.0.0/16", "2001:0db8::/32"},
		internalCIDRs:  []string{"10.3.0.0/16", "2001:0db8::/32"},
		internalRanges: []LoadBalancer_IPRange{},
	},
	{
		name:           "IPv4Range",
		apiCIDRs:       []string{"10.3.0.10-10.3.0.20"},
		internalCIDRs:  []string{},
		internalRanges: []LoadBalancer_IPRange{{Start: "10.3.0.10", Stop: "10.3.0.20"}},
	},
	{
		name:           "IPv4CIDRAndRange",
		apiCIDRs:       []string{"10.3.0.32/28", "10.3.0.10-10.3.0.20"},
		internalCIDRs:  []string{"10.3.0.32/28"},
		internalRanges: []LoadBalancer_IPRange{{Start: "10.3.0.10", Stop: "10.3.0.20"}},
	},
	{
		name:           "IPv6Range",
		apiCIDRs:       []string{"2001:0db8::0-2001:0db8::10"},
		internalCIDRs:  []string{},
		internalRanges: []LoadBalancer_IPRange{{Start: "2001:0db8::0", Stop: "2001:0db8::10"}},
	},
	{
		name:           "IPv4CIDRAndIPv6Range",
		apiCIDRs:       []string{"10.3.0.32/28", "2001:0db8::0-2001:0db8::10"},
		internalCIDRs:  []string{"10.3.0.32/28"},
		internalRanges: []LoadBalancer_IPRange{{Start: "2001:0db8::0", Stop: "2001:0db8::10"}},
	},
	{name: "Empty", apiCIDRs: []string{""}, expectErr: true},
	{name: "Invalid", apiCIDRs: []string{"bananas"}, expectErr: true},
	{name: "CommaSeparated", apiCIDRs: []string{"fd01::/64,fd02::/64,fd03::/64"}, expectErr: true},
	{name: "InvalidStartIPv4", apiCIDRs: []string{"10.3.0.1000-10.3.0.1001"}, expectErr: true},
	{name: "InvalidStopIPv4", apiCIDRs: []string{"10.3.0.10-10.3.0.300"}, expectErr: true},
	{name: "Order", apiCIDRs: []string{"10.3.0.10-10.3.0.7"}, expectErr: true},
	{name: "InvalidStopIPv6", apiCIDRs: []string{"2001:0db8::0-2001:0db8::gg"}, expectErr: true},
	{name: "AnyFail", apiCIDRs: []string{"", "10.3.0.7-10.3.0.10"}, expectErr: true},
	{name: "InvalidRange", apiCIDRs: []string{"", "10.3.0.10-10.3.0.12-10.3.0.15"}, expectErr: true},
}

func Test_loadBalancerCIDRsFromAPI(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		g := NewWithT(t)
		cidrs, ranges, err := loadBalancerCIDRsFromAPI(nil)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(cidrs).To(BeNil())
		g.Expect(ranges).To(BeNil())
	})

	for _, tc := range loadBalancerCIDRTestCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			cidrs, ranges, err := loadBalancerCIDRsFromAPI(&tc.apiCIDRs)
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(*cidrs).To(Equal(tc.internalCIDRs))
				g.Expect(*ranges).To(Equal(tc.internalRanges))
			}
		})
	}
}

func Test_loadBalancerCIDRsToAPI(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(loadBalancerCIDRsToAPI(nil, nil)).To(BeNil())
		})
		t.Run("CIDRsOnly", func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(loadBalancerCIDRsToAPI(nil, utils.Pointer([]LoadBalancer_IPRange{}))).To(Equal(utils.Pointer([]string{})))
		})
		t.Run("RangesOnly", func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(loadBalancerCIDRsToAPI(utils.Pointer([]string{}), nil)).To(Equal(utils.Pointer([]string{})))
		})
	})

	for _, tc := range loadBalancerCIDRTestCases {
		if !tc.expectErr {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)

				cidrs := loadBalancerCIDRsToAPI(&tc.internalCIDRs, &tc.internalRanges)
				g.Expect(*cidrs).To(Equal(tc.apiCIDRs))
			})
		}
	}
}
