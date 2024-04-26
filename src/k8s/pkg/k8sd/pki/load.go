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
	if keyPEM == "" {
		return cert, nil, nil
	}

	key, err := LoadRSAPrivateKey(keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load RSA private key: %w", err)
	}

	return cert, key, nil
}

// LoadRSAPrivateKey parses the specified PEM block and return the rsa.PrivateKey.
func LoadRSAPrivateKey(keyPEM string) (*rsa.PrivateKey, error) {
	pb, _ := pem.Decode([]byte(keyPEM))
	if pb == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	switch pb.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(pb.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
		}
		return key, nil
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(pb.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		v, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
		return v, nil
	}
	return nil, fmt.Errorf("unknown private key block type %q", pb.Type)
}

// LoadRSAPublicKey parses the specified PEM block and return the rsa.PublicKey.
func LoadRSAPublicKey(keyPEM string) (*rsa.PublicKey, error) {
	pb, _ := pem.Decode([]byte(keyPEM))
	if pb == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	switch pb.Type {
	case "PUBLIC KEY":
		parsed, err := x509.ParsePKIXPublicKey(pb.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		v, ok := parsed.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA public key")
		}
		return v, nil
	}
	return nil, fmt.Errorf("unknown public key block type %q", pb.Type)
}
