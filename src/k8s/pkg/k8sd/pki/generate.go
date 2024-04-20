package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
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

// generateSerialNumber returns a random number that can be used for the SerialNumber field in an x509 certificate.
func generateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	return serialNumber, nil
}

func generateCertificate(subject pkix.Name, years int, ca bool, dnsSANs []string, ipSANs []net.IP) (*x509.Certificate, error) {
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for certificate template: %w", err)
	}

	cert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(years, 0, 0),
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

func generateSelfSignedCA(subject pkix.Name, years int, bits int) (string, string, error) {
	cert, err := generateCertificate(subject, years, true, nil, nil)
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

func signCertificate(certificate *x509.Certificate, bits int, parent *x509.Certificate, pub any, priv any) (string, string, error) {
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

func generateRSAKey(bits int) (string, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", fmt.Errorf("failed to generate RSA private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if keyPEM == nil {
		return "", fmt.Errorf("failed to encode private key PEM")
	}
	return string(keyPEM), nil
}

func generateECDSAKeypair(curve elliptic.Curve) (string, string, error) {
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ECDSA private key: %w", err)
	}
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode ECDSA private key: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	if privPEM == nil {
		return "", "", fmt.Errorf("failed to encode private key PEM")
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode ECDSA public key: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	if pubPEM == nil {
		return "", "", fmt.Errorf("failed to encode public key PEM")
	}

	return string(privPEM), string(pubPEM), nil
}
