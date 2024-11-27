package pki_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

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

// patchCertPEM can be used to modify certificates for testing purposes.
func patchCertPEM(
	certPEM string,
	caPEM string,
	caKeyPEM string,
	updateFunc func(*x509.Certificate) error,
) (string, string, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return "", "", fmt.Errorf("failed to decode certificate")
	}

	cert, _ := x509.ParseCertificate(block.Bytes)
	if cert == nil {
		return "", "", fmt.Errorf("failed to decode certificate")
	}

	// Generate a new certificate based on the input certificate and the
	// updates applied by "updateFunc".
	template, err := pkiutil.GenerateCertificate(
		cert.Subject,
		cert.NotBefore, cert.NotAfter, false,
		cert.DNSNames, cert.IPAddresses,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate patched certificate")
	}

	if err = updateFunc(template); err != nil {
		return "", "", fmt.Errorf("cert update failed: %w", err)
	}

	caCert, caKey, err := pkiutil.LoadCertificate(caPEM, caKeyPEM)
	if err != nil {
		return "", "", fmt.Errorf("failed to load CA cert: %w", err)
	}

	certPem, keyPem, err := pkiutil.SignCertificate(template, 2048, caCert, &caCert.PublicKey, caKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign cert: %w", err)
	}

	return certPem, keyPem, err
}

func TestControlPlaneCertificates(t *testing.T) {
	notBefore := time.Now()
	opts := pki.ControlPlanePKIOpts{
		Hostname:          "h1",
		NotBefore:         notBefore,
		NotAfter:          notBefore.AddDate(1, 0, 0),
		AllowSelfSignedCA: true,
	}
	c := pki.NewControlPlanePKI(opts)

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
			Hostname:  "h1",
			NotBefore: notBefore,
			NotAfter:  notBefore.AddDate(1, 0, 0),
		})

		c.CACert = mustReadTestData(t, "ca.pem")

		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).ToNot(Succeed())
	})

	t.Run("ApiServerCertSANs", func(t *testing.T) {
		c := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{
			Hostname:          "h1",
			NotBefore:         notBefore,
			NotAfter:          notBefore.AddDate(1, 0, 0),
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
			NotBefore:         notBefore,
			NotAfter:          notBefore.AddDate(1, 0, 0),
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

	t.Run("InvalidSAN", func(t *testing.T) {
		c := pki.NewControlPlanePKI(opts)
		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).To(Succeed())

		// Switch CA certificates, expecting certificate validation failures.
		c.CACert = c.FrontProxyCACert
		c.CAKey = c.FrontProxyCAKey
		g.Expect(c.CompleteCertificates()).ToNot(Succeed())
	})

	t.Run("KubeletCertExpired", func(t *testing.T) {
		c := pki.NewControlPlanePKI(opts)
		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).To(Succeed())

		var err error
		c.KubeletCert, c.KubeletKey, err = patchCertPEM(c.KubeletCert, c.CACert, c.CAKey, func(cert *x509.Certificate) error {
			cert.NotAfter = time.Now().AddDate(-1, 0, 0)
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		err = c.CompleteCertificates()
		g.Expect(err).To(MatchError(ContainSubstring("certificate expired")))
	})

	t.Run("KubeletCertNotBefore", func(t *testing.T) {
		c := pki.NewControlPlanePKI(opts)
		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).To(Succeed())

		var err error
		c.KubeletCert, c.KubeletKey, err = patchCertPEM(c.KubeletCert, c.CACert, c.CAKey, func(cert *x509.Certificate) error {
			cert.NotBefore = time.Now().AddDate(1, 0, 0)
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		err = c.CompleteCertificates()
		g.Expect(err).To(MatchError(ContainSubstring("invalid certificate, not valid before")))
	})

	t.Run("KubeletCertInvalidCN", func(t *testing.T) {
		c := pki.NewControlPlanePKI(opts)
		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).To(Succeed())

		var err error
		c.KubeletCert, c.KubeletKey, err = patchCertPEM(c.KubeletCert, c.CACert, c.CAKey, func(cert *x509.Certificate) error {
			cert.Subject = pkix.Name{
				CommonName:   "unexpected-cn",
				Organization: cert.Subject.Organization,
			}
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		err = c.CompleteCertificates()
		g.Expect(err).To(MatchError(ContainSubstring("invalid certificate CN")))
	})

	t.Run("KubeletCertInvalidOrganization", func(t *testing.T) {
		c := pki.NewControlPlanePKI(opts)
		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).To(Succeed())

		var err error
		c.KubeletCert, c.KubeletKey, err = patchCertPEM(c.KubeletCert, c.CACert, c.CAKey, func(cert *x509.Certificate) error {
			cert.Subject = pkix.Name{
				CommonName:   cert.Subject.CommonName,
				Organization: []string{"unexpected-organization"},
			}
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		err = c.CompleteCertificates()
		g.Expect(err).To(MatchError(ContainSubstring("missing cert organization")))
	})

	t.Run("KubeletCertInvalidDNSName", func(t *testing.T) {
		c := pki.NewControlPlanePKI(opts)
		g := NewWithT(t)
		g.Expect(c.CompleteCertificates()).To(Succeed())

		var err error
		c.KubeletCert, c.KubeletKey, err = patchCertPEM(c.KubeletCert, c.CACert, c.CAKey, func(cert *x509.Certificate) error {
			cert.DNSNames = []string{"some-other-dnsname"}
			return nil
		})
		g.Expect(err).ToNot(HaveOccurred())

		err = c.CompleteCertificates()
		g.Expect(err).To(MatchError(MatchRegexp(`certificate dns name \(.*\) validation failure`)))
	})
}
