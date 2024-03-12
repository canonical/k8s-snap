package pki

import (
	"crypto/x509/pkix"
	"fmt"
	"net"
)

// K8sDqlitePKI is a list of certificates required by the k8s-dqlite datastore.
type K8sDqlitePKI struct {
	allowSelfSignedCA bool     // create self-signed CA certificates if missing
	hostname          string   // node name
	ipSANs            []net.IP // IP SANs for generated certificates
	dnsSANs           []string // DNS SANs for the certificates below
	years             int      // how many years the generated certificates will be valid for

	// CN=k8s-dqlite, DNS=hostname, IP=127.0.0.1 (self-signed)
	K8sDqliteCert, K8sDqliteKey string
}

type K8sDqlitePKIOpts struct {
	Hostname          string
	DNSSANs           []string
	IPSANs            []net.IP
	Years             int
	AllowSelfSignedCA bool
	Datastore         string
}

func NewK8sDqlitePKI(opts K8sDqlitePKIOpts) *K8sDqlitePKI {
	if opts.Years == 0 {
		opts.Years = 1
	}

	return &K8sDqlitePKI{
		allowSelfSignedCA: opts.AllowSelfSignedCA,
		hostname:          opts.Hostname,
		years:             opts.Years,
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

		template, err := generateCertificate(pkix.Name{CommonName: "k8s"}, c.years, false, append(c.dnsSANs, c.hostname), append(c.ipSANs, net.IP{127, 0, 0, 1}))
		if err != nil {
			return fmt.Errorf("failed to generate k8s-dqlite certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, template, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to self-sign k8s-dqlite certificate: %w", err)
		}

		c.K8sDqliteCert = cert
		c.K8sDqliteKey = key
	}

	return nil
}
