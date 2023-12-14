package certutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"path/filepath"
	"time"
)

// CertificateTemplate represents a certificate that will be generated.
type CertificateTemplate struct {
	// The x509 certificate to sign.
	cert *x509.Certificate
	// The size of RSA key that will be generated.
	pkiBits int
	// The path where the generated private key and certificate will reside.
	certFolder string
	// The name of the .key and .crt files.
	certName string
}

// NewCertificateTemplate returns a certificate template that can be signed later on.
// For input it takes:
//   - subject (the subject, e.g. containing CN field for the certificate)
//   - dnsNames (used in subjectAltName)
//   - ipAddresses (used in subjectAltName)
//   - expiry (the expiry date of the certificate, e.g. for a certificate valid for 10 years would be time.Now().AddDate(10, 0, 0))
//   - pkiBits (the bit size of the RSA private key)
//   - certFolder (the folder where the generated private key and certificate will reside)
//   - certName (the name that will be used to create the .key and .crt files)
func NewCertificateTemplate(subject pkix.Name, dnsNames []string, ipAddresses []net.IP, expiry time.Time, pkiBits int, certFolder, certName string) (*CertificateTemplate, error) {
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for certificate template: %w", err)
	}

	return &CertificateTemplate{
		cert: &x509.Certificate{
			SerialNumber:          serialNumber,
			Subject:               subject,
			NotBefore:             time.Now(),
			NotAfter:              expiry,
			IsCA:                  false,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature,
			BasicConstraintsValid: true,
			DNSNames:              dnsNames,
			IPAddresses:           ipAddresses,
		},
		pkiBits:    pkiBits,
		certFolder: certFolder,
		certName:   certName,
	}, nil
}

// NewCaTemplate returns a certificate template that will create a certificate authority.
// For input it takes:
//   - subject
//   - expiry (the expiry date of the certificate, e.g. for a certificate valid for 10 years would be time.Now().AddDate(10, 0, 0))
//   - pkiBits (the bit size of the RSA private key)
//   - certFolder (the folder where the generated private key and certificate will reside)
//   - certName (the name that will be used to create the .key and .crt files)
func NewCATemplate(subject pkix.Name, expiry time.Time, pkiBits int, certFolder, certName string) (*CertificateTemplate, error) {
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for certificate template: %w", err)
	}

	return &CertificateTemplate{
		cert: &x509.Certificate{
			SerialNumber:          serialNumber,
			Subject:               subject,
			NotBefore:             time.Now(),
			NotAfter:              expiry,
			IsCA:                  true,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
		},
		pkiBits:    pkiBits,
		certFolder: certFolder,
		certName:   certName,
	}, nil
}

// SignAndSaveCertificate creates, signs and saves a certificate from the template.
func (ctpl *CertificateTemplate) SignAndSave(selfSign bool, ca *CertKeyPair) (*CertKeyPair, error) {
	key, err := rsa.GenerateKey(rand.Reader, ctpl.pkiBits)
	if err != nil {
		return nil, fmt.Errorf("failed to create RSA private key: %w", err)
	}

	ckp, err := NewCertKeyPair(ctpl.cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert-key pair: %w", err)
	}

	err = ckp.Sign(selfSign, ca)
	if err != nil {
		return nil, fmt.Errorf("failed to sign certificate: %w", err)
	}

	ckp.SavePrivateKey(filepath.Join(ctpl.certFolder, fmt.Sprintf("%s.key", ctpl.certName)))
	if err != nil {
		return nil, fmt.Errorf("failed to save private key: %w", err)
	}

	ckp.SaveCertificate(filepath.Join(ctpl.certFolder, fmt.Sprintf("%s.crt", ctpl.certName)))
	if err != nil {
		return nil, fmt.Errorf("failed to save certificate: %w", err)
	}

	return ckp, nil
}
