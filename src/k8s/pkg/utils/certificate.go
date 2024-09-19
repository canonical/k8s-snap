package utils

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"time"
)

// SplitIPAndDNSSANs splits a list of SANs into IP and DNS SANs
// Returns a list of IP addresses and a list of DNS names
func SplitIPAndDNSSANs(extraSANs []string) ([]net.IP, []string) {
	var ipSANs []net.IP
	var dnsSANs []string

	for _, san := range extraSANs {
		if san == "" {
			continue
		}

		if ip := net.ParseIP(san); ip != nil {
			ipSANs = append(ipSANs, ip)
		} else {
			dnsSANs = append(dnsSANs, san)
		}
	}

	return ipSANs, dnsSANs
}

// TLSClientConfig returns a TLS configuration that trusts a remote server
// The remoteCert is the public key of the server we are connecting to.
// The rootCAs is the list of trusted CAs, allowing you to pass the clients existing trusted CAs.
func TLSClientConfigWithTrustedCertificate(remoteCert *x509.Certificate, rootCAs *x509.CertPool) (*tls.Config, error) {
	config := &tls.Config{}
	if remoteCert == nil {
		return nil, fmt.Errorf("invalid remote public key")
	}

	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	config.RootCAs = rootCAs
	remoteCert.IsCA = true
	config.RootCAs.AddCert(remoteCert)

	// Always use public key DNS name rather than server cert, so that it matches.
	if len(remoteCert.DNSNames) > 0 {
		config.ServerName = remoteCert.DNSNames[0]
	}

	return config, nil
}

// GetRemoteCertificate retrieves the remote certificate from a given address
// The address should be in the format of "hostname:port"
// Returns the remote certificate or an error
func GetRemoteCertificate(address string) (*x509.Certificate, error) {
	// validate address
	_, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("failed to validate the cluster member address: %w", err)
	}

	url := fmt.Sprintf("https://%s", address)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Connect
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Retrieve the certificate
	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return nil, fmt.Errorf("unable to read remote TLS certificate")
	}

	return resp.TLS.PeerCertificates[0], nil
}

// CertFingerprint returns the SHA256 fingerprint of a certificate
func CertFingerprint(cert *x509.Certificate) string {
	return fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
}

// GetCertExpiry returns the expiration date of a certificate in RFC3339 format.
func GetCertExpiry(certPEM string) (string, error) {
	// Decode the PEM certificate
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse certificate PEM")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Return the expiration date formatted as RFC3339
	return cert.NotAfter.Format(time.RFC3339), nil
}
