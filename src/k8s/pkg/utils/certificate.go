package utils

import (
	"crypto"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
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

func CreateHTTPClientWithCert(cert *x509.Certificate, key crypto.PrivateKey) (*http.Client, error) {
	// Create the client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cert.Raw},
				PrivateKey:  key,
			},
		},
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

	// Connect
	req, err := http.NewRequest("GET", address, nil)
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
