package pkiutil

import (
	"crypto/x509"
	"fmt"
	"slices"
	"time"
)

// CertCheck can be used to validate certificates. Unspecified fields are
// ignored. "NotBefore" and "NotAfter" are checked implicitly.
type CertCheck struct {
	// Ensure that the certificate has the specified Common Name.
	CN string
	// Ensure that the certificate contains the following organizations.
	O []string
	// Ensure that the certificate contains the following DNS SANs.
	DNSSANs []string
	// Validate the certificate against the specified CA certificate.
	CaPEM           string
	AllowSelfSigned bool
}

func (check CertCheck) ValidateKeypair(certPEM string, keyPEM string) error {
	cert, _, err := LoadCertificate(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	return check.ValidateCert(cert)
}

func (check CertCheck) ValidateCert(cert *x509.Certificate) error {
	if check.CN != "" && check.CN != cert.Subject.CommonName {
		return fmt.Errorf("invalid certificate CN, expected: %s, actual: %s ",
			check.CN, cert.Subject.CommonName)
	}
	for _, checkO := range check.O {
		if !slices.Contains(cert.Subject.Organization, checkO) {
			return fmt.Errorf("missing cert organization: %s, actual: %v",
				checkO, cert.Subject.Organization)
		}
	}

	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("invalid certificate, not valid before: %v, current time: %v",
			cert.NotBefore, now)
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate expired since: %v, current time: %v",
			cert.NotAfter, now)
	}

	if !check.AllowSelfSigned {
		verifyOpts := x509.VerifyOptions{}
		if check.CaPEM != "" {
			roots := x509.NewCertPool()
			if !roots.AppendCertsFromPEM([]byte(check.CaPEM)) {
				return fmt.Errorf("invalid CA certificate")
			}
			verifyOpts.Roots = roots
		}

		if _, err := cert.Verify(verifyOpts); err != nil {
			return fmt.Errorf("certificate validation failure: %w", err)
		}
	}

	for _, dnsName := range check.DNSSANs {
		if err := cert.VerifyHostname(dnsName); err != nil {
			return fmt.Errorf("certificate dns name (%s) validation failure: %w, allowed dns names: %v",
				dnsName, err, cert.DNSNames)
		}
	}

	return nil
}
