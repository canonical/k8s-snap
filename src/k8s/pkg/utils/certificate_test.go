package utils_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/utils"

	. "github.com/onsi/gomega"
)

func TestSplitIPAndDNSSANs(t *testing.T) {
	tests := []string{"192.168.0.1", "::1", "cluster.local", "kubernetes.svc.local", "", "2001:db8:0:1:1:1:1:1"}

	g := NewWithT(t)
	gotIPs, gotDNSs := utils.SplitIPAndDNSSANs(tests)

	// Convert cert.IPAddresses to a slice of string representations
	ips := make([]string, len(gotIPs))
	for i, ip := range gotIPs {
		ips[i] = ip.String()
	}

	g.Expect(len(ips)).To(Equal(3))
	g.Expect(ips).To(ConsistOf("192.168.0.1", "::1", "2001:db8:0:1:1:1:1:1"))

	g.Expect(len(gotDNSs)).To(Equal(2))
	g.Expect(gotDNSs).To(ConsistOf("cluster.local", "kubernetes.svc.local"))
}
