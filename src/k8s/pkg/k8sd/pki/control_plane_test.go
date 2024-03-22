package pki_test

import (
	"crypto/x509"
	"embed"
	"encoding/pem"
	"io/fs"
	"net"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	. "github.com/onsi/gomega"
)

//go:embed data
var testCertificates embed.FS

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

		cert, err := fs.ReadFile(testCertificates, "data/ca.pem")
		g.Expect(err).To(BeNil())
		c.CACert = string(cert)

		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).ToNot(BeNil())
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
		g.Expect(c.CompleteCertificates()).To(BeNil())

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

			for _, expectedIP := range expectedIPs {
				t.Run(expectedIP, func(t *testing.T) {
					g.Expect(actualIPs).To(ContainElement(expectedIP), "IP should be present: "+expectedIP)
				})
			}

			g.Expect(cert.IPAddresses).To(HaveLen(len(expectedIPs)))
		})

		t.Run("DNSNames", func(t *testing.T) {
			g := NewWithT(t)
			expectedDNSNames := []string{"cluster.local", "kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster", "kubernetes.default.svc.cluster.local"}

			for _, expectedDNS := range expectedDNSNames {
				t.Run(expectedDNS, func(t *testing.T) {
					g.Expect(cert.DNSNames).To(ContainElement(expectedDNS), "DNS should be present: "+expectedDNS)
				})
			}

			g.Expect(cert.DNSNames).To(HaveLen(len(expectedDNSNames)))
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
		g.Expect(c.CompleteCertificates()).To(BeNil())

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

			for _, expectedIP := range expectedIPs {
				t.Run(expectedIP, func(t *testing.T) {
					g.Expect(actualIPs).To(ContainElement(expectedIP), "IP should be present: "+expectedIP)
				})
			}

			g.Expect(cert.IPAddresses).To(HaveLen(len(expectedIPs)))
		})

		t.Run("DNSNames", func(t *testing.T) {
			g := NewWithT(t)
			expectedDNSNames := []string{"h1", "cluster.local"}

			for _, expectedDNS := range expectedDNSNames {
				t.Run(expectedDNS, func(t *testing.T) {
					g.Expect(cert.DNSNames).To(ContainElement(expectedDNS), "DNS should be present: "+expectedDNS)
				})
			}

			g.Expect(cert.DNSNames).To(HaveLen(len(expectedDNSNames)))
		})
	})
}
