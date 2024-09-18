package pki

import (
	"crypto/x509/pkix"
	"fmt"
	"net"
	"time"

	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
)

// K8sDqlitePKI is a list of certificates required by the k8s-dqlite datastore.
type K8sDqlitePKI struct {
	allowSelfSignedCA bool      // create self-signed CA certificates if missing
	hostname          string    // node name
	ipSANs            []net.IP  // IP SANs for generated certificates
	dnsSANs           []string  // DNS SANs for the certificates below
	notBefore         time.Time // notBefore date for the generated certificates
	notAfter          time.Time // not after date (expiration date) for the generated certificates

	// CN=k8s, DNS=hostname, IP=127.0.0.1 (self-signed)
	K8sDqliteCert, K8sDqliteKey string
}

type K8sDqlitePKIOpts struct {
	Hostname          string
	DNSSANs           []string
	IPSANs            []net.IP
	NotBefore         time.Time
	NotAfter          time.Time
	AllowSelfSignedCA bool
	Datastore         string
}

func NewK8sDqlitePKI(opts K8sDqlitePKIOpts) *K8sDqlitePKI {
	// NOTE: Default NotAfter is 1 year from the NotBefore date
	if opts.NotAfter.IsZero() {
		opts.NotAfter = opts.NotBefore.AddDate(1, 0, 0)
	}

	return &K8sDqlitePKI{
		allowSelfSignedCA: opts.AllowSelfSignedCA,
		hostname:          opts.Hostname,
		notBefore:         opts.NotBefore,
		notAfter:          opts.NotAfter,
		ipSANs:            opts.IPSANs,
		dnsSANs:           opts.DNSSANs,
	}
}

// CompleteCertificates generates missing or unset certificates. If only a certificate is set and not a key, we assume that the cluster is using managed certificates.
func (c *K8sDqlitePKI) CompleteCertificates() error {
	// Fail hard if keys of self-signed certificates are set without the respective certificates
	switch {
	case c.K8sDqliteCert == "" && c.K8sDqliteKey != "":
		return fmt.Errorf("k8s-dqlite certificate key set without a certificate, fail to prevent further issues")
	case c.K8sDqliteCert != "" && c.K8sDqliteKey == "":
		return fmt.Errorf("k8s-dqlite certificate set without a key, fail to prevent further issues")
	}

	// Generate k8s-dqlite client certificate (if missing)
	if c.K8sDqliteCert == "" && c.K8sDqliteKey == "" {
		if !c.allowSelfSignedCA {
			return fmt.Errorf("k8s-dqlite certificate not specified and generating self-signed certificates is not allowed")
		}

		template, err := pkiutil.GenerateCertificate(pkix.Name{CommonName: "k8s"}, c.notBefore, c.notAfter, false, append(c.dnsSANs, c.hostname), append(c.ipSANs, net.ParseIP("127.0.0.1"), net.ParseIP("::1")))
		if err != nil {
			return fmt.Errorf("failed to generate k8s-dqlite certificate: %w", err)
		}
		cert, key, err := pkiutil.SignCertificate(template, 2048, template, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to self-sign k8s-dqlite certificate: %w", err)
		}

		c.K8sDqliteCert = cert
		c.K8sDqliteKey = key
	}

	return nil
}
