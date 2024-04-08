package pki

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// loadCertificate parses the PEM blocks and returns the certificate and private key.
// loadCertificate will fail if certPEM is not a valid certificate.
// loadCertificate will return a nil private key if keyPEM is empty, but will fail if it is not valid.
func loadCertificate(certPEM string, keyPEM string) (*x509.Certificate, *rsa.PrivateKey, error) {
	decodedCert, _ := pem.Decode([]byte(certPEM))
	if decodedCert == nil {
		return nil, nil, fmt.Errorf("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(decodedCert.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	var key *rsa.PrivateKey
	if keyPEM != "" {
		pb, _ := pem.Decode([]byte(keyPEM))
		switch pb.Type {
		case "RSA PRIVATE KEY":
			key, err = x509.ParsePKCS1PrivateKey(pb.Bytes)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse RSA private key: %w", err)
			}
		case "PRIVATE KEY":
			parsed, err := x509.ParsePKCS8PrivateKey(pb.Bytes)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
			}
			v, ok := parsed.(*rsa.PrivateKey)
			if !ok {
				return nil, nil, fmt.Errorf("not an RSA private key")
			}
			key = v
		default:
			return nil, nil, fmt.Errorf("unknown private key block type %q", pb.Type)
		}
	}
	return cert, key, nil
}
