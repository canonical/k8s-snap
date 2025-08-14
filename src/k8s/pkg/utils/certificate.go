package utils

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"
)

// SplitIPAndDNSSANs splits a list of SANs into IP and DNS SANs
// Returns a list of IP addresses and a list of DNS names.
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

// Options for TLS handshake checking.
type TLSCheckOptions struct {
	Timeout              time.Duration
	InsecureSkipVerify   bool
	ClientCertSkipVerify bool
	ClientCertFile       string
	ClientKeyFile        string
}

func TLSHandshakeCheck(address string, opts TLSCheckOptions) (*tls.ConnectionState, error) {
	// Validates that the address is an active TLS server
	// even if it requires a client certificate, it will return without error.

	_, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("failed to validate the cluster member address: %w", err)
	}

	var serverRequestedCert bool

	var certs []tls.Certificate
	if opts.ClientCertFile != "" && opts.ClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(opts.ClientCertFile, opts.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %w", err)
		}
		opts.ClientCertSkipVerify = false
		certs = []tls.Certificate{cert}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: opts.InsecureSkipVerify,
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			serverRequestedCert = true
			return &certs[0], nil
		},
	}

	// Create a dialer with a timeout
	dialer := &net.Dialer{Timeout: opts.Timeout}
	// Perform the TLS handshake
	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	if err != nil {
		if serverRequestedCert && opts.ClientCertSkipVerify {
			// Server requested a client cert, but we are skipping client cert verification
			// so we ignore this error.
			return nil, nil
		}
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	return &state, nil
}

// GetRemoteCertificate retrieves the remote certificate from a given address
// The address should be in the format of "hostname:port"
// Returns the remote certificate or an error.
func GetRemoteCertificate(address string) (*x509.Certificate, error) {
	conn_state, err := TLSHandshakeCheck(address, TLSCheckOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check TLS: %w", err)
	}

	// Retrieve the certificate
	if conn_state == nil || len(conn_state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("unable to read remote TLS certificate")
	}

	return conn_state.PeerCertificates[0], nil
}

// CertFingerprint returns the SHA256 fingerprint of a certificate.
func CertFingerprint(cert *x509.Certificate) string {
	return fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
}
