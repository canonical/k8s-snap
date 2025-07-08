package pki

import (
	"crypto/x509/pkix"
	"fmt"
	"net"
	"time"

	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
)

// EtcdPKI is a list of certificates required by the etcd datastore.
type EtcdPKI struct {
	allowSelfSignedCA bool      // create self-signed CA certificates if missing
	hostname          string    // node name
	ipSANs            []net.IP  // IP SANs for generated certificates
	dnsSANs           []string  // DNS SANs for the certificates below
	notBefore         time.Time // notBefore date for the generated certificates
	notAfter          time.Time // not after date (expiration date) for the generated certificates

	// CN=etcd, DNS=hostname, IP=127.0.0.1 (self-signed)
	CACert, CAKey string

	// [server] CN=kube-etcd, DNS=hostname, IP=127.0.0.1,address (signed by etcd-ca)
	ServerCert, ServerKey string

	// [server] CN=kube-etcd-peer, DNS=hostname, IP=127.0.0.1,address (signed by etcd-ca)
	ServerPeerCert, ServerPeerKey string

	// [client] CN=kube-apiserver-etcd-client (signed by etcd-ca)
	APIServerClientCert, APIServerClientKey string
}

type EtcdPKIOpts struct {
	Hostname          string
	DNSSANs           []string
	IPSANs            []net.IP
	NotBefore         time.Time
	NotAfter          time.Time
	AllowSelfSignedCA bool
}

func NewEtcdPKI(opts EtcdPKIOpts) *EtcdPKI {
	// NOTE: Default NotAfter is 1 year from the NotBefore date
	if opts.NotAfter.IsZero() {
		opts.NotAfter = opts.NotBefore.AddDate(1, 0, 0)
	}

	return &EtcdPKI{
		allowSelfSignedCA: opts.AllowSelfSignedCA,
		hostname:          opts.Hostname,
		notAfter:          opts.NotAfter,
		notBefore:         opts.NotBefore,
		ipSANs:            opts.IPSANs,
		dnsSANs:           opts.DNSSANs,
	}
}

// CompleteCertificates generates missing or unset certificates. If only a certificate is set and not a key, we assume that the cluster is using managed certificates.
func (c *EtcdPKI) CompleteCertificates() error {
	// Fail hard if keys of self-signed certificates are set without the respective certificates
	if c.CACert == "" && c.CAKey != "" {
		return fmt.Errorf("etcd CA certificate key set without a certificate, fail to prevent further issues")
	}

	// Generate self-signed CA (if not set already)
	if c.CACert == "" && c.CAKey == "" {
		if !c.allowSelfSignedCA {
			return fmt.Errorf("etcd CA not specified and generating self-signed CA not allowed")
		}
		cert, key, err := pkiutil.GenerateSelfSignedCA(pkix.Name{CommonName: "etcd-ca"}, c.notBefore, c.notAfter, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate etcd CA: %w", err)
		}
		c.CACert = cert
		c.CAKey = key
	} else {
		certCheck := pkiutil.CertCheck{AllowSelfSigned: true}
		if err := certCheck.ValidateKeypair(c.CACert, c.CAKey); err != nil {
			return fmt.Errorf("etcd CA certificate validation failure: %w", err)
		}
	}

	cert, key, err := pkiutil.LoadCertificate(c.CACert, c.CAKey)
	if err != nil {
		return fmt.Errorf("failed to parse etcd CA: %w", err)
	}

	// Generate etcd server certificate
	if c.ServerCert == "" && c.ServerKey == "" {
		if key == nil {
			return fmt.Errorf("using an external etcd CA with specifying an etcd server certificate is not possible")
		}
		template, err := pkiutil.GenerateCertificate(pkix.Name{CommonName: "kube-etcd"}, c.notBefore, c.notAfter, false, append(c.dnsSANs, c.hostname), append(c.ipSANs, net.IP{127, 0, 0, 1}))
		if err != nil {
			return fmt.Errorf("failed to generate etcd certificate: %w", err)
		}
		cert, key, err := pkiutil.SignCertificate(template, 2048, cert, &key.PublicKey, key)
		if err != nil {
			return fmt.Errorf("failed to self-sign etcd certificate: %w", err)
		}

		c.ServerCert = cert
		c.ServerKey = key
	}

	// Generate etcd peer server certificate
	if c.ServerPeerCert == "" && c.ServerPeerKey == "" {
		if key == nil {
			return fmt.Errorf("using an external etcd CA with specifying an etcd server peer certificate is not possible")
		}

		template, err := pkiutil.GenerateCertificate(pkix.Name{CommonName: "kube-etcd-peer"}, c.notBefore, c.notAfter, false, append(c.dnsSANs, c.hostname), append(c.ipSANs, net.IP{127, 0, 0, 1}))
		if err != nil {
			return fmt.Errorf("failed to generate etcd certificate: %w", err)
		}
		cert, key, err := pkiutil.SignCertificate(template, 2048, cert, &key.PublicKey, key)
		if err != nil {
			return fmt.Errorf("failed to self-sign etcd certificate: %w", err)
		}

		c.ServerPeerCert = cert
		c.ServerPeerKey = key
	}

	// Generate kube-apiserver etcd client certificate
	if c.APIServerClientCert == "" && c.APIServerClientKey == "" {
		if key == nil {
			return fmt.Errorf("using an external etcd CA with specifying an etcd apiserver client certificate is not possible")
		}

		template, err := pkiutil.GenerateCertificate(pkix.Name{CommonName: "kube-apiserver-etcd-client"}, c.notBefore, c.notAfter, false, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to generate etcd certificate: %w", err)
		}
		cert, key, err := pkiutil.SignCertificate(template, 2048, cert, &key.PublicKey, key)
		if err != nil {
			return fmt.Errorf("failed to self-sign etcd certificate: %w", err)
		}

		c.APIServerClientCert = cert
		c.APIServerClientKey = key
	}

	return nil
}
