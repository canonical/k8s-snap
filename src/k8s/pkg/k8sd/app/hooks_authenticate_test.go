package app

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func generateTestCert(t *testing.T, notBefore, notAfter time.Time) *x509.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}
	return cert
}

func TestOnAuthenticate(t *testing.T) {
	now := time.Now()
	validCert := generateTestCert(t, now.Add(-time.Hour), now.Add(time.Hour))
	otherCert := generateTestCert(t, now.Add(-time.Hour), now.Add(time.Hour))
	expiredCert := generateTestCert(t, now.Add(-2*time.Hour), now.Add(-time.Hour))
	futureCert := generateTestCert(t, now.Add(time.Hour), now.Add(2*time.Hour))

	trusted := map[string]x509.Certificate{
		"valid":   *validCert,
		"expired": *expiredCert,
		"future":  *futureCert,
	}

	for _, tc := range []struct {
		name        string
		tlsState    *tls.ConnectionState
		expectTrust bool
		expectErr   bool
	}{
		{name: "NoTLS", tlsState: nil, expectErr: true},
		{name: "NoPeerCertificates", tlsState: &tls.ConnectionState{}, expectErr: true},
		{name: "NotYetValid", tlsState: &tls.ConnectionState{PeerCertificates: []*x509.Certificate{futureCert}}, expectErr: true},
		{name: "Expired", tlsState: &tls.ConnectionState{PeerCertificates: []*x509.Certificate{expiredCert}}, expectErr: true},
		{name: "Untrusted", tlsState: &tls.ConnectionState{PeerCertificates: []*x509.Certificate{otherCert}}},
		{name: "Trusted", tlsState: &tls.ConnectionState{PeerCertificates: []*x509.Certificate{validCert}}, expectTrust: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.TLS = tc.tlsState

			a := &App{}
			trustedResult, err := a.onAuthenticate(context.Background(), nil, r, trusted)
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
			g.Expect(trustedResult).To(Equal(tc.expectTrust))
		})
	}
}
