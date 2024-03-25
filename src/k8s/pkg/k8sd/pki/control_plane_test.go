package pki_test

import (
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	. "github.com/onsi/gomega"
)

func mustReadTestData(t *testing.T, filename string) string {
	data, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestControlPlaneCertificates(t *testing.T) {
	c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
		Hostname:          "h1",
		Years:             10,
		AllowSelfSignedCA: true,
	})

	g := NewWithT(t)

	g.Expect(c.CompleteCertificates()).To(BeNil())
	g.Expect(c.CompleteCertificates()).To(BeNil())

	t.Run("MissingCAKey", func(t *testing.T) {
		c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
			Hostname: "h1",
			Years:    10,
		})

		c.CACert = mustReadTestData(t, "ca.pem")

		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).ToNot(Succeed())
	})

	t.Run("ApiServerCertSANs", func(t *testing.T) {
		c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
			Hostname:          "h1",
			Years:             10,
			AllowSelfSignedCA: true,
			IPSANs:            []net.IP{net.ParseIP("192.168.2.123")},
			DNSSANs:           []string{"cluster.local"},
		})

		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).To(Succeed())

		block, _ := pem.Decode([]byte(c.APIServerCert))
		g.Expect(block).ToNot(BeNil())

		cert, _ := x509.ParseCertificate(block.Bytes)
		g.Expect(cert).ToNot(BeNil())

		t.Run("IPAddresses", func(t *testing.T) {
			g := NewWithT(t)
			expectedIPs := []string{"192.168.2.123", "127.0.0.1", "::1"}

			// Convert cert.IPAddresses to a slice of string representations
			actualIPs := make([]string, len(cert.IPAddresses))
			for i, ip := range cert.IPAddresses {
				actualIPs[i] = ip.String()
			}

			g.Expect(actualIPs).To(ConsistOf(expectedIPs))
		})

		t.Run("DNSNames", func(t *testing.T) {
			g := NewWithT(t)
			expectedDNSNames := []string{"cluster.local", "kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster", "kubernetes.default.svc.cluster.local"}

			g.Expect(cert.DNSNames).To(ConsistOf(expectedDNSNames))
		})
	})

	t.Run("KubeletCertSANs", func(t *testing.T) {
		c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
			Hostname:          "h1",
			Years:             10,
			AllowSelfSignedCA: true,
			IPSANs:            []net.IP{net.ParseIP("192.168.2.123")},
			DNSSANs:           []string{"cluster.local"},
		})

		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).To(Succeed())

		block, _ := pem.Decode([]byte(c.KubeletCert))
		g.Expect(block).ToNot(BeNil())

		cert, _ := x509.ParseCertificate(block.Bytes)
		g.Expect(cert).ToNot(BeNil())

		t.Run("IPAddresses", func(t *testing.T) {
			g := NewWithT(t)
			expectedIPs := []string{"192.168.2.123", "127.0.0.1", "::1"}

			// Convert cert.IPAddresses to a slice of string representations
			actualIPs := make([]string, len(cert.IPAddresses))
			for i, ip := range cert.IPAddresses {
				actualIPs[i] = ip.String()
			}

			g.Expect(actualIPs).To(ConsistOf(expectedIPs))
		})

		t.Run("DNSNames", func(t *testing.T) {
			g := NewWithT(t)
			expectedDNSNames := []string{"h1", "cluster.local"}

			g.Expect(cert.DNSNames).To(ConsistOf(expectedDNSNames))
		})
	})
}
