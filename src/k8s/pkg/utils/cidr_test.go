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
