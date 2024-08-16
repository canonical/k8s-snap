package pkiutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// GenerateSerialNumber returns a random number that can be used for the SerialNumber field in an x509 certificate.
func GenerateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	return serialNumber, nil
}

func GenerateCertificate(subject pkix.Name, seconds int, ca bool, dnsSANs []string, ipSANs []net.IP) (*x509.Certificate, error) {
	serialNumber, err := GenerateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for certificate template: %w", err)
	}

	cert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(seconds) * time.Second),
		IPAddresses:           ipSANs,
		DNSNames:              dnsSANs,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}
	if ca {
		cert.IsCA = true
		cert.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign
	} else {
		cert.IsCA = false
		cert.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature
	}

	return cert, nil
}

func GenerateSelfSignedCA(subject pkix.Name, seconds int, bits int) (string, string, error) {
	cert, err := GenerateCertificate(subject, seconds, true, nil, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate certificate: %w", err)
	}

	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if keyPEM == nil {
		return "", "", fmt.Errorf("failed to encode private key PEM")
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		return "", "", fmt.Errorf("failed to self-sign certificate: %w", err)
	}
	crtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if crtPEM == nil {
		return "", "", fmt.Errorf("failed to encode certificate PEM")
	}

	return string(crtPEM), string(keyPEM), nil
}

func SignCertificate(certificate *x509.Certificate, bits int, parent *x509.Certificate, pub any, priv any) (string, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if keyPEM == nil {
		return "", "", fmt.Errorf("failed to encode private key PEM")
	}

	if pub == nil && priv == nil {
		priv = key
		pub = &key.PublicKey
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, certificate, parent, &key.PublicKey, priv)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign certificate: %w", err)
	}
	crtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if crtPEM == nil {
		return "", "", fmt.Errorf("failed to encode certificate PEM")
	}

	return string(crtPEM), string(keyPEM), nil
}

func GenerateRSAKey(bits int) (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA private key: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if privPEM == nil {
		return "", "", fmt.Errorf("failed to encode private key PEM")
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode RSA public key: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	if pubPEM == nil {
		return "", "", fmt.Errorf("failed to encode public key PEM")
	}

	return string(privPEM), string(pubPEM), nil
}

// GenerateCSR generates a certificate signing request (CSR) and private key for the given subject.
func GenerateCSR(subject pkix.Name, bits int, dnsSANs []string, ipSANs []net.IP) (string, string, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if keyPEM == nil {
		return "", "", fmt.Errorf("failed to encode private key PEM")
	}

	csrTemplate := &x509.CertificateRequest{
		Subject:     subject,
		DNSNames:    dnsSANs,
		IPAddresses: ipSANs,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, key)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate request: %w", err)
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	if csrPEM == nil {
		return "", "", fmt.Errorf("failed to encode certificate request PEM")
	}

	return string(csrPEM), string(keyPEM), nil
}
