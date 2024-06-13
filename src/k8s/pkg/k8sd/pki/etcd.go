package pki

import (
	"crypto/x509/pkix"
	"fmt"
	"net"
)

// EtcdPKI is a list of certificates required by the embedded etcd datastore.
type EtcdPKI struct {
	allowSelfSignedCA bool     // create self-signed CA certificates if missing
	hostname          string   // node name
	ipSANs            []net.IP // IP SANs for generated certificates
	dnsSANs           []string // DNS SANs for the certificates below
	years             int      // how many years the generated certificates will be valid for

	// CN=k8s-dqlite, DNS=hostname, IP=127.0.0.1 (self-signed)
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
	Years             int
	AllowSelfSignedCA bool
}

func NewEtcdPKI(opts EtcdPKIOpts) *EtcdPKI {
	if opts.Years == 0 {
		opts.Years = 10
	}

	return &EtcdPKI{
		allowSelfSignedCA: opts.AllowSelfSignedCA,
		hostname:          opts.Hostname,
		years:             opts.Years,
		ipSANs:            opts.IPSANs,
		dnsSANs:           opts.DNSSANs,
	}
}

// CompleteCertificates generates missing or unset certificates. If only a certificate is set and not a key, we assume that the cluster is using managed certificates.
func (c *EtcdPKI) CompleteCertificates() error {
	// Fail hard if keys of self-signed certificates are set without the respective certificates
	switch {
	case c.CACert == "" && c.CAKey != "":
		return fmt.Errorf("etcd CA certificate key set without a certificate, fail to prevent further issues")
	case c.CACert != "" && c.CAKey == "":
		return fmt.Errorf("etcd CA certificate set without a key, fail to prevent further issues")
	}

	// Generate self-signed CA (if not set already)
	if c.CACert == "" && c.CAKey == "" {
		if !c.allowSelfSignedCA {
			return fmt.Errorf("etcd CA not specified and generating self-signed CA not allowed")
		}
		cert, key, err := generateSelfSignedCA(pkix.Name{CommonName: "etcd-ca"}, c.years, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate etcd CA: %w", err)
		}
		c.CACert = cert
		c.CAKey = key
	}

	cert, key, err := loadCertificate(c.CACert, c.CAKey)
	if err != nil {
		return fmt.Errorf("failed to parse etcd CA: %w", err)
	}

	// Generate etcd server certificate
	if c.ServerCert == "" && c.ServerKey == "" {
		if key == nil {
			return fmt.Errorf("using an external etcd CA with specifying an etcd server certificate is not possible")
		}
		template, err := generateCertificate(pkix.Name{CommonName: "kube-etcd"}, c.years, false, append(c.dnsSANs, c.hostname), append(c.ipSANs, net.IP{127, 0, 0, 1}))
		if err != nil {
			return fmt.Errorf("failed to generate k8s-dqlite certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, cert, &key.PublicKey, key)
		if err != nil {
			return fmt.Errorf("failed to self-sign k8s-dqlite certificate: %w", err)
		}

		c.ServerCert = cert
		c.ServerKey = key
	}

	// Generate etcd peer server certificate
	if c.ServerPeerCert == "" && c.ServerPeerKey == "" {
		if key == nil {
			return fmt.Errorf("using an external etcd CA with specifying an etcd server peer certificate is not possible")
		}

		template, err := generateCertificate(pkix.Name{CommonName: "kube-etcd-peer"}, c.years, false, append(c.dnsSANs, c.hostname), append(c.ipSANs, net.IP{127, 0, 0, 1}))
		if err != nil {
			return fmt.Errorf("failed to generate k8s-dqlite certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, cert, &key.PublicKey, key)
		if err != nil {
			return fmt.Errorf("failed to self-sign k8s-dqlite certificate: %w", err)
		}

		c.ServerPeerCert = cert
		c.ServerPeerKey = key
	}

	// Generate kube-apiserver etcd client certificate
	if c.APIServerClientCert == "" && c.APIServerClientKey == "" {
		if key == nil {
			return fmt.Errorf("using an external etcd CA with specifying an etcd apiserver client certificate is not possible")
		}

		template, err := generateCertificate(pkix.Name{CommonName: "kube-apiserver-etcd-client"}, c.years, false, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to generate k8s-dqlite certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, cert, &key.PublicKey, key)
		if err != nil {
			return fmt.Errorf("failed to self-sign k8s-dqlite certificate: %w", err)
		}

		c.APIServerClientCert = cert
		c.APIServerClientKey = key
	}

	return nil
}
