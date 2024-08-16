package pki_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
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
		Seconds:           3600,
		AllowSelfSignedCA: true,
	})

	g := NewWithT(t)

	g.Expect(c.CompleteCertificates()).To(Succeed())
	g.Expect(c.CompleteCertificates()).To(Succeed())

	t.Run("K8sdKey", func(t *testing.T) {
		g := NewWithT(t)

		priv, err := pkiutil.LoadRSAPrivateKey(c.K8sdPrivateKey)
		g.Expect(err).ToNot(HaveOccurred())
		pub, err := pkiutil.LoadRSAPublicKey(c.K8sdPublicKey)
		g.Expect(err).ToNot(HaveOccurred())

		// generate a hash to sign
		b := make([]byte, 10)
		rand.Read(b)
		h := sha256.New()
		h.Write(b)
		hashed := h.Sum(nil)

		// sign hash
		signed, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed)
		g.Expect(err).ToNot(HaveOccurred())

		// verify signature
		g.Expect(rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed, signed)).To(Succeed())
	})

	t.Run("MissingCAKey", func(t *testing.T) {
		c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
			Hostname: "h1",
			Seconds:  3600,
		})

		c.CACert = mustReadTestData(t, "ca.pem")

		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).ToNot(Succeed())
	})

	t.Run("ApiServerCertSANs", func(t *testing.T) {
		c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
			Hostname:          "h1",
			Seconds:           3600,
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
			Seconds:           3600,
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
