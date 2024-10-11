package utils_test

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
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

	g.Expect(ips).To(ConsistOf("192.168.0.1", "::1", "2001:db8:0:1:1:1:1:1"))

	g.Expect(gotDNSs).To(ConsistOf("cluster.local", "kubernetes.svc.local"))
}

func TestTLSClientConfigWithTrustedCertificate(t *testing.T) {
	g := NewWithT(t)

	// Mock certificate and certificate pool for testing
	remoteCert := &x509.Certificate{
		DNSNames: []string{"bubblegum.com"},
	}
	rootCAs := x509.NewCertPool()

	tlsConfig, err := utils.TLSClientConfigWithTrustedCertificate(remoteCert, rootCAs)

	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(tlsConfig.ServerName).To(Equal("bubblegum.com"))
	g.Expect(tlsConfig.RootCAs.Subjects()).To(ContainElement(remoteCert.RawSubject))

	// Test with invalid remote certificate
	tlsConfig, err = utils.TLSClientConfigWithTrustedCertificate(nil, rootCAs)
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(tlsConfig).To(BeNil())

	// Test with nil root CAs
	_, err = utils.TLSClientConfigWithTrustedCertificate(remoteCert, nil)
	g.Expect(err).To(Not(HaveOccurred()))
}

func TestCertFingerprint(t *testing.T) {
	g := NewWithT(t)
	// Create a mock certificate for testing
	mockCert := &x509.Certificate{
		Raw: []byte("ChocolateChipCookieDough"),
	}

	// Calculate the expected SHA256 fingerprint of the mock certificate
	expectedFingerprint := sha256.Sum256(mockCert.Raw)
	expectedFingerprintStr := hex.EncodeToString(expectedFingerprint[:])

	// Call the CertFingerprint function
	actualFingerprint := utils.CertFingerprint(mockCert)

	// Check if the returned fingerprint matches the expected fingerprint
	g.Expect(actualFingerprint).To(Equal(expectedFingerprintStr))
}

func TestGetRemoteCertificate(t *testing.T) {
	g := NewWithT(t)
	// Create a mock HTTP server that returns a mock certificate
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test with a valid address
	remoteCert, err := utils.GetRemoteCertificate(server.Listener.Addr().String())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(remoteCert).To(Equal(server.Certificate()))

	// Test with an invalid address (missing port)
	remoteCert, err = utils.GetRemoteCertificate("candy.canes")
	g.Expect(err).To(HaveOccurred())
	g.Expect(remoteCert).To(BeNil())

	// Test with a non-existent address
	remoteCert, err = utils.GetRemoteCertificate("jellybeans:9999")
	g.Expect(err).To(HaveOccurred())
	g.Expect(remoteCert).To(BeNil())
}
