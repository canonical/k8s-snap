package types

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_loadBalancerCIDRsFromAPI(t *testing.T) {
	for _, tc := range []struct {
		name         string
		cidrs        []string
		expectCIDRs  []string
		expectRanges []LoadBalancer_IPRange
		expectErr    bool
	}{
		{},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			cidrs, ranges, err := loadBalancerCIDRsFromAPI(tc.cidrs)
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(cidrs).To(Equal(tc.expectCIDRs))
				g.Expect(ranges).To(Equal(tc.expectRanges))
			}
		})
	}

}
