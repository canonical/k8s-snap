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

	g.Expect(err).To(BeNil())
	g.Expect(tlsConfig.ServerName).To(Equal("bubblegum.com"))
	g.Expect(tlsConfig.RootCAs.Subjects()).To(ContainElement(remoteCert.RawSubject))

	// Test with invalid remote certificate
	tlsConfig, err = utils.TLSClientConfigWithTrustedCertificate(nil, rootCAs)
	g.Expect(err).ToNot(BeNil())
	g.Expect(tlsConfig).To(BeNil())

	// Test with nil root CAs
	_, err = utils.TLSClientConfigWithTrustedCertificate(remoteCert, nil)
	g.Expect(err).To(BeNil())
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

func TestGetCertExpiry(t *testing.T) {
	certPEM := `-----BEGIN CERTIFICATE-----
MIIDkzCCAnugAwIBAgIUHjmfFK9cfwsJ9wVeay77DUxGfsQwDQYJKoZIhvcNAQEL
BQAwWTELMAkGA1UEBhMCVVMxDjAMBgNVBAgMBVN0YXRlMQ0wCwYDVQQHDARDaXR5
MRUwEwYDVQQKDAxPcmdhbml6YXRpb24xFDASBgNVBAMMC2V4YW1wbGUuY29tMB4X
DTI0MDMyMTE5MTE1NVoXDTM0MDMxOTE5MTE1NVowWTELMAkGA1UEBhMCVVMxDjAM
BgNVBAgMBVN0YXRlMQ0wCwYDVQQHDARDaXR5MRUwEwYDVQQKDAxPcmdhbml6YXRp
b24xFDASBgNVBAMMC2V4YW1wbGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAqiI/f/IWAk+2I3uxoTxB20RxrwUvPglAmsvXpkT40PbCHZ9pqI4I
GQoY5mR4bQMx4s3TQNIMGIIha9IvLXVgQb6WxZNc7lLWOg/VHAw+0tUkGnO2o89v
loRNJj2+ZcFu9UZQDLa/cr5pKGnFI4O3rR8DcQxt9rPtSY62ICLFwqU2Hw3fjyHI
FITKmTrZNccmcWKBuOfj4DkFaFT9+jZ72W8DHBXMjAm7qZC3ar9ZlzhHT8mI942i
LuNd0r47yrzga/kLCtjHDYXjBGBareIsfAZDJ+1WV9wVShL42brTwchZhBVcxY66
by8PZJPD97c22zvVyCKIUGGcFKxvWb2fBQIDAQABo1MwUTAdBgNVHQ4EFgQU3LTT
fZ/8wUZhUj856yEniIkE6xwwHwYDVR0jBBgwFoAU3LTTfZ/8wUZhUj856yEniIkE
6xwwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEABRNgwMKqA5Y8
7wfa+X3RsoG0BVOF/+GYCtyXBwH3lXzlOrkTbkL4e9rYGmPx67VCpsnCEAhipta3
FqjGyZFhMsaaIDlhJjm+K7MTGA7aSfo6NIBmpPRKjIQFL2rhmqs1r7riafwvvDrU
CzhIi7rODCf7NAzoISU1EzowzKdKNgGYMNvIpv1pMd7p7WHQNK+W+gvQJZ93UpDY
o9fgMdo44Am9bsiiPi7LAWU5qzbdUErrgFslI+inwD3dOxIwBGfEfD0ngz2nF+Jh
S63GKldmH7KYVE4sdB2BvfgiraDTTHRIDNre930YIhVI+XLHIhtJ+BSpFO4w/idC
xjvgVUetag==
-----END CERTIFICATE-----`

	testCases := []struct {
		name          string
		input         string
		expected      string
		expectedError bool
	}{
		{
			name:     "Valid certificate",
			input:    certPEM,
			expected: "2034-03-19T19:11:55Z",
		},
		{
			name:          "Invalid certificate content",
			input:         "-----BEGIN CERTIFICATE-----\ninvalid\n-----END CERTIFICATE-----",
			expectedError: true,
		},
		{
			name:          "Empty PEM block",
			input:         "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			actual, err := utils.GetCertExpiry(tc.input)
			if tc.expectedError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(actual).To(Equal(tc.expected))
			}
		})
	}
}
