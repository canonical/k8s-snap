package pki

import (
	"crypto/x509/pkix"
	"fmt"
	"net"
)

// ControlPlanePKI is a list of all certificates we require for a control plane node.
type ControlPlanePKI struct {
	allowSelfSignedCA bool     // create self-signed CA certificates if missing
	hostname          string   // node name
	ipSANs            []net.IP // IP SANs for generated certificates
	dnsSANs           []string // DNS SANs for the certificates below
	years             int      // how many years the generated certificates will be valid for

	CACert, CAKey                             string // CN=kubernetes-ca (self-signed)
	FrontProxyCACert, FrontProxyCAKey         string // CN=kubernetes-front-proxy-ca (self-signed)
	FrontProxyClientCert, FrontProxyClientKey string // CN=front-proxy-client (signed by kubernetes-front-proxy-ca)
	ServiceAccountKey                         string // private key used to sign service account tokens

	// CN=k8s-dqlite, DNS=hostname, IP=127.0.0.1 (self-signed)
	K8sDqliteCert, K8sDqliteKey string

	// CN=kube-apiserver, DNS=hostname,kubernetes.* IP=127.0.0.1,10.152.183.1,address (signed by kubernetes-ca)
	APIServerCert, APIServerKey string

	// CN=kube-apiserver-kubelet-client, O=system:masters (signed by kubernetes-ca)
	APIServerKubeletClientCert, APIServerKubeletClientKey string

	// CN=system:node:hostname, O=system:nodes, DNS=hostname, IP=127.0.0.1,address (signed by kubernetes-ca)
	KubeletCert, KubeletKey string
}

type ControlPlanePKIOpts struct {
	Hostname          string
	DNSSANs           []string
	IPSANs            []net.IP
	Years             int
	AllowSelfSignedCA bool
}

func NewControlPlanePKI(opts ControlPlanePKIOpts) *ControlPlanePKI {
	if opts.Years == 0 {
		opts.Years = 1
	}

	return &ControlPlanePKI{
		allowSelfSignedCA: opts.AllowSelfSignedCA,
		hostname:          opts.Hostname,
		years:             opts.Years,
		ipSANs:            opts.IPSANs,
		dnsSANs:           opts.DNSSANs,
	}
}

// CompleteCertificates generates missing or unset certificates. If only a certificate is set and not a key, we assume that the cluster is using managed certificates.
func (c *ControlPlanePKI) CompleteCertificates() error {
	// Fail hard if keys of self-signed certificates are set without the respective certificates
	switch {
	case c.CACert == "" && c.CAKey != "":
		return fmt.Errorf("kubernetes CA key is set without a certificate, fail to prevent causing issues")
	case c.FrontProxyCACert == "" && c.FrontProxyCAKey != "":
		return fmt.Errorf("front-proxy CA key is set without a certificate, fail to prevent causing issues")
	case c.K8sDqliteCert == "" && c.K8sDqliteKey != "":
		return fmt.Errorf("k8s-dqlite certificate key set without a certificate, fail to prevent further issues")
	case c.K8sDqliteCert != "" && c.K8sDqliteKey == "":
		return fmt.Errorf("k8s-dqlite certificate set without a key, fail to prevent further issues")
	}

	// Generate self-signed CA (if not set already)
	if c.CACert == "" && c.CAKey == "" {
		if !c.allowSelfSignedCA {
			return fmt.Errorf("kubernetes CA not specified and generating self-signed CA not allowed")
		}
		cert, key, err := generateSelfSignedCA(pkix.Name{CommonName: "kubernetes-ca"}, c.years, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate kubernetes CA: %w", err)
		}
		c.CACert = cert
		c.CAKey = key
	}

	caCertificate, caPrivateKey, err := loadCertificate(c.CACert, c.CAKey)
	if err != nil {
		return fmt.Errorf("failed to parse kubernetes CA: %w", err)
	}

	// Generate self-signed CA for front-proxy (if not set already)
	if c.FrontProxyCACert == "" && c.FrontProxyCAKey == "" {
		if !c.allowSelfSignedCA {
			return fmt.Errorf("front-proxy CA not specified and generating self-signed CA not allowed")
		}
		cert, key, err := generateSelfSignedCA(pkix.Name{CommonName: "front-proxy-ca"}, c.years, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate front-proxy CA: %w", err)
		}
		c.FrontProxyCACert = cert
		c.FrontProxyCAKey = key
	}

	// Generate front proxy client certificate (ok to override)
	if c.FrontProxyClientCert == "" || c.FrontProxyClientKey == "" {
		frontProxyCACert, frontProxyCAKey, err := loadCertificate(c.FrontProxyCACert, c.FrontProxyCAKey)
		switch {
		case err != nil:
			return fmt.Errorf("failed to parse front proxy CA: %w", err)
		case frontProxyCAKey == nil:
			return fmt.Errorf("using an external front proxy CA without providing the front-proxy-client certificate is not possible")
		}

		template, err := generateCertificate(pkix.Name{CommonName: "front-proxy-client"}, c.years, false, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to generate front-proxy-client certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, frontProxyCACert, &frontProxyCAKey.PublicKey, frontProxyCAKey)
		if err != nil {
			return fmt.Errorf("failed to sign front-proxy-client certificate: %w", err)
		}

		c.FrontProxyClientCert = cert
		c.FrontProxyClientKey = key
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

	// Generate service account key (if missing)
	if c.ServiceAccountKey == "" {
		if !c.allowSelfSignedCA {
			return fmt.Errorf("service account signing key not specified and generating new key is not allowed")
		}

		key, err := generateKey(2048)
		if err != nil {
			return fmt.Errorf("failed to generate service account key: %w", err)
		}

		c.ServiceAccountKey = key
	}

	// Generate kubelet certificate (if missing)
	if c.KubeletCert == "" || c.KubeletKey == "" {
		if caPrivateKey == nil {
			return fmt.Errorf("using an external kubernetes CA without providing the kubelet certificate is not possible")
		}

		template, err := generateCertificate(
			pkix.Name{CommonName: fmt.Sprintf("system:node:%s", c.hostname), Organization: []string{"system:nodes"}},
			c.years, false, append(c.dnsSANs, c.hostname), append(c.ipSANs, net.IP{127, 0, 0, 1}),
		)
		if err != nil {
			return fmt.Errorf("failed to generate kubelet certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, caCertificate, &caPrivateKey.PublicKey, caPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to sign kubelet certificate: %w", err)
		}

		c.KubeletCert = cert
		c.KubeletKey = key
	}

	// Generate apiserver-kubelet-client certificate (if missing)
	if c.APIServerKubeletClientCert == "" || c.APIServerKubeletClientKey == "" {
		if caPrivateKey == nil {
			return fmt.Errorf("using an external kubernetes CA without providing the apiserver-kubelet-client certificate is not possible")
		}

		template, err := generateCertificate(pkix.Name{CommonName: "apiserver-kubelet-client", Organization: []string{"system:masters"}}, c.years, false, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to generate apiserver-kubelet-client certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, caCertificate, &caPrivateKey.PublicKey, caPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to sign apiserver-kubelet-client certificate: %w", err)
		}

		c.APIServerKubeletClientCert = cert
		c.APIServerKubeletClientKey = key
	}

	// Generate kube-apiserver certificate (if missing)
	if c.APIServerCert == "" || c.APIServerKey == "" {
		if caPrivateKey == nil {
			return fmt.Errorf("using an external kubernetes CA without providing the apiserver certificate is not possible")
		}

		// TODO(neoaggelos): we also need to specify the kubernetes service IP here, not hardcode 10.152.183.1
		template, err := generateCertificate(
			pkix.Name{CommonName: "kube-apiserver"},
			c.years,
			false,
			append(c.dnsSANs, "kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster", "kubernetes.default.svc.cluster.local"), append(c.ipSANs, net.IP{10, 152, 183, 1}, net.IP{127, 0, 0, 1}))
		if err != nil {
			return fmt.Errorf("failed to generate apiserver certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, caCertificate, &caPrivateKey.PublicKey, caPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to sign apiserver certificate: %w", err)
		}

		c.APIServerCert = cert
		c.APIServerKey = key
	}
	return nil
}
