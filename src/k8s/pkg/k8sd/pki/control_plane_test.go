package pki_test

import (
	"crypto/ecdsa"
	"crypto/rand"
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

	g.Expect(c.CompleteCertificates()).To(Succeed())
	g.Expect(c.CompleteCertificates()).To(Succeed())

	t.Run("K8sdKey", func(t *testing.T) {
		g := NewWithT(t)

		privBlock, _ := pem.Decode([]byte(c.K8sdPrivateKey))
		priv, err := x509.ParseECPrivateKey(privBlock.Bytes)
		g.Expect(err).To(Succeed())

		b := make([]byte, 10)
		_, err = rand.Read(b)
		g.Expect(err).To(Succeed())

		signed, err := ecdsa.SignASN1(rand.Reader, priv, b)
		g.Expect(err).To(Succeed())

		pubBlock, _ := pem.Decode([]byte(c.K8sdPublicKey))
		pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
		g.Expect(err).To(Succeed())

		pub, ok := pubKey.(*ecdsa.PublicKey)
		g.Expect(ok).To(BeTrue())

		g.Expect(ecdsa.VerifyASN1(pub, b, signed)).To(BeTrue())
	})

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
			expectedIPs := []net.IP{net.ParseIP("192.168.2.123").To4(), net.ParseIP("127.0.0.1").To4(), net.ParseIP("::1")}

			g.Expect(cert.IPAddresses).To(ConsistOf(expectedIPs))
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
			expectedIPs := []net.IP{net.ParseIP("192.168.2.123").To4(), net.ParseIP("127.0.0.1").To4(), net.ParseIP("::1")}

			g.Expect(cert.IPAddresses).To(ConsistOf(expectedIPs))
		})

		t.Run("DNSNames", func(t *testing.T) {
			g := NewWithT(t)
			expectedDNSNames := []string{"h1", "cluster.local"}

			g.Expect(cert.DNSNames).To(ConsistOf(expectedDNSNames))
		})
	})
}
