package utils

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
)

// SplitIPAndDNSSANs splits a list of SANs into IP and DNS SANs
// Returns a list of IP addresses and a list of DNS names
func SplitIPAndDNSSANs(extraSANs []string) ([]net.IP, []string) {
	var ipSANs []net.IP
	var dnsSANs []string

	for _, san := range extraSANs {
		if san == "" {
			log.Println("Skipping empty SAN")
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

func TLSClientConfig(remoteCert *x509.Certificate) (*tls.Config, error) {
	if remoteCert == nil {
		return nil, fmt.Errorf("invalid remote public key")
	}

	config := &tls.Config{}

	// Add the public key to the CA pool to make it trusted.
	config.RootCAs = x509.NewCertPool()
	remoteCert.IsCA = true
	remoteCert.KeyUsage = x509.KeyUsageCertSign
	config.RootCAs.AddCert(remoteCert)

	// Always use public key DNS name rather than server cert, so that it matches.
	if len(remoteCert.DNSNames) > 0 {
		config.ServerName = remoteCert.DNSNames[0]
	}

	return config, nil
}

func CreateHTTPClientWithCert(cert *x509.Certificate) (*http.Client, error) {
	// Create the client
	tlsConfig, err := TLSClientConfig(cert)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}, nil
}

func GetRemoteCertificate(address string) (*x509.Certificate, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	url, err := url.Parse("https://" + address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address: %w", err)
	}

	// Connect
	req, err := http.NewRequest("GET", url.String(), nil)
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

func CertFingerprint(cert *x509.Certificate) string {
	return fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
}
