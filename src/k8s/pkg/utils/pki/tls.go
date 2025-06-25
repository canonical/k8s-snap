package pkiutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// LoadTLSConfigFromPath loads TLS certificates from the given file paths.
func LoadTLSConfigFromPath(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate or key: %w", err)
	}

	caCertPool, err := LoadCACertPool(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}, nil
}

// LoadCACertPool loads the CA certificate pool from the given file.
func LoadCACertPool(caFile string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to append CA certificate")
	}
	return caCertPool, nil
}
