package pki_test

import (
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

// TestControlPlaneEndpointCertSAN asserts that a configured ControlPlaneEndpoint host lands in
// the kube-apiserver serving certificate SANs, wired exactly as the bootstrap/join/refresh hooks
// do it (endpoint -> SANs() -> ControlPlanePKIOpts -> cert).
func TestControlPlaneEndpointCertSAN(t *testing.T) {
	for _, tc := range []struct {
		name      string
		host      string
		expectIP  string
		expectDNS string
	}{
		{name: "IPv4", host: "10.0.0.250", expectIP: "10.0.0.250"},
		{name: "DNS", host: "api.example.com", expectDNS: "api.example.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			endpoint := types.ControlPlaneEndpoint{Host: utils.Pointer(tc.host)}
			endpointIPs, endpointNames := endpoint.SANs()

			notBefore := time.Now()
			c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
				Hostname:          "h1",
				NotBefore:         notBefore,
				NotAfter:          notBefore.AddDate(1, 0, 0),
				AllowSelfSignedCA: true,
				IPSANs:            append([]net.IP{net.ParseIP("192.168.2.123")}, endpointIPs...),
				DNSSANs:           append([]string{"cluster.local"}, endpointNames...),
			})
			g.Expect(c.CompleteCertificates()).To(Succeed())

			block, _ := pem.Decode([]byte(c.APIServerCert))
			g.Expect(block).ToNot(BeNil())
			cert, err := x509.ParseCertificate(block.Bytes)
			g.Expect(err).ToNot(HaveOccurred())

			if tc.expectIP != "" {
				ips := make([]string, 0, len(cert.IPAddresses))
				for _, ip := range cert.IPAddresses {
					ips = append(ips, ip.String())
				}
				g.Expect(ips).To(ContainElement(tc.expectIP))
			}
			if tc.expectDNS != "" {
				g.Expect(cert.DNSNames).To(ContainElement(tc.expectDNS))
			}
		})
	}
}
