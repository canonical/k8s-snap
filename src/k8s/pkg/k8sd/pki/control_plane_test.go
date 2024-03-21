package pki_test

import (
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	data "github.com/canonical/k8s/pkg/k8sd/pki/data"
	. "github.com/onsi/gomega"
)

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

		c.CACert = data.DUMMY_CA_CERT

		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).ToNot(BeNil())
	})

	t.Run("ExtraSANs", func(t *testing.T) {
		c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
			Hostname:          "h1",
			Years:             10,
			AllowSelfSignedCA: true,
			ExtraSANs:         "192.168.2.123,cluster.local",
		})

		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).To(BeNil())

		block, _ := pem.Decode([]byte(c.APIServerCert))
		g.Expect(block).ToNot(BeNil())

		cert, _ := x509.ParseCertificate(block.Bytes)
		g.Expect(cert).ToNot(BeNil())

		t.Run("IPAddresses", func(t *testing.T) {
			g := NewWithT(t)
			found := false
			for _, ip := range cert.IPAddresses {
				if ip.Equal(net.ParseIP("192.168.2.123")) {
					found = true
					break
				}
			}
			g.Expect(found).To(BeTrue())
		})

		t.Run("DNSNames", func(t *testing.T) {
			g := NewWithT(t)
			found := false
			for _, dns := range cert.DNSNames {
				if dns == "cluster.local" {
					found = true
					break
				}
			}
			g.Expect(found).To(BeTrue())
		})
	})
}
