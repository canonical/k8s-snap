package certutils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// CertKeyPair represents a private key and x509 certificate pair.
type CertKeyPair struct {
	// Key is the generated RSA private key.
	Key *rsa.PrivateKey
	// Cert is the x509 certificate.
	Cert *x509.Certificate
	// KeyPem is the key in PEM format.
	KeyPem []byte
	// CertPem is the certificate in PEM format.
	CertPem []byte
}

// Sign signs the x509 certificate with the private key.
func (ckp *CertKeyPair) Sign(selfSign bool, ca *CertKeyPair) (err error) {
	var derBytes []byte

	if selfSign {
		derBytes, err = x509.CreateCertificate(rand.Reader, ckp.Cert, ckp.Cert, &ckp.Key.PublicKey, ckp.Key)

	} else {
		derBytes, err = x509.CreateCertificate(rand.Reader, ckp.Cert, ca.Cert, &ckp.Key.PublicKey, ca.Key)
	}
	if err != nil {
		return fmt.Errorf("failed to create x509 certificate: %w", err)
	}

	certOut := &bytes.Buffer{}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return err
	}
	ckp.CertPem = certOut.Bytes()

	return nil
}

// NewCertKeyPair returns a new pair from provided certificate and private key.
func NewCertKeyPair(cert *x509.Certificate, privateKey *rsa.PrivateKey) (*CertKeyPair, error) {
	ckp := &CertKeyPair{}
	ckp.Key = privateKey
	ckp.Cert = cert

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	keyOut := &bytes.Buffer{}
	err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return nil, err
	}
	ckp.KeyPem = keyOut.Bytes()

	return ckp, nil
}

// SavePrivateKey marshals the key in PKCS1 format, encodes it as a PEM block and saves it to the given path.
func (ckp *CertKeyPair) SavePrivateKey(path string) error {
	keyFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	_, err = keyFile.Write(ckp.KeyPem)
	if err != nil {
		return err
	}

	return nil
}

// SaveCertificate encodes the given certificate as a PEM block and saves it to the given path.
func (ckp *CertKeyPair) SaveCertificate(path string) error {
	certFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer certFile.Close()

	_, err = certFile.Write(ckp.CertPem)
	if err != nil {
		return err
	}

	return nil
}

// LoadCertKeyPair loads a key and certificate pair from given paths.
func LoadCertKeyPair(keyPath string, certPath string) (*CertKeyPair, error) {
	ckp := &CertKeyPair{}

	dat, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	ckp.KeyPem = dat

	pb, _ := pem.Decode(dat)
	key, err := x509.ParsePKCS1PrivateKey(pb.Bytes)
	if err != nil {
		return nil, err
	}
	ckp.Key = key

	dat, err = os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	ckp.CertPem = dat

	pb, _ = pem.Decode(dat)

	cert, err := x509.ParseCertificate(pb.Bytes)
	if err != nil {
		return nil, err
	}
	ckp.Cert = cert

	return ckp, nil
}
