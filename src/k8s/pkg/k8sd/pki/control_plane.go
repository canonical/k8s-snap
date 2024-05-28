package pki

import (
	"crypto/x509/pkix"
	"fmt"
	"net"
)

// ControlPlanePKI is a list of all certificates we require for a control plane node.
type ControlPlanePKI struct {
	allowSelfSignedCA         bool     // create self-signed CA certificates if missing
	includeMachineAddressSANs bool     // include any machine IP addresses as SANs for generated certificates
	hostname                  string   // node name
	ipSANs                    []net.IP // IP SANs for generated certificates
	dnsSANs                   []string // DNS SANs for the certificates below
	years                     int      // how many years the generated certificates will be valid for

	CACert, CAKey                             string // CN=kubernetes-ca (self-signed)
	ClientCACert, ClientCAKey                 string // CN=kubernetes-ca-client (self-signed)
	FrontProxyCACert, FrontProxyCAKey         string // CN=kubernetes-front-proxy-ca (self-signed)
	FrontProxyClientCert, FrontProxyClientKey string // CN=front-proxy-client (signed by kubernetes-front-proxy-ca)
	ServiceAccountKey                         string // private key used to sign service account tokens

	// [server] CN=kube-apiserver, DNS=hostname,kubernetes.* IP=127.0.0.1,10.152.183.1,address (signed by kubernetes-ca)
	APIServerCert, APIServerKey string

	// [server] CN=system:node:hostname, O=system:nodes, DNS=hostname, IP=127.0.0.1,address (signed by kubernetes-ca)
	KubeletCert, KubeletKey string

	// [client] CN=kubernetes:admin, O=system:masters (signed by kubernetes-ca-client)
	AdminClientCert, AdminClientKey string

	// [client] CN=system:kube-controller-manager (signed by kubernetes-ca-client)
	KubeControllerManagerClientCert, KubeControllerManagerClientKey string

	// [client] CN=system:kube-scheduler (signed by kubernetes-ca-client)
	KubeSchedulerClientCert, KubeSchedulerClientKey string

	// [client] CN=system:kube-proxy (signed by kubernetes-ca-client)
	KubeProxyClientCert, KubeProxyClientKey string

	// [client] CN=system:node:hostname, O=system:nodes (signed by kubernetes-ca-client)
	KubeletClientCert, KubeletClientKey string

	// [client] CN=kube-apiserver-kubelet-client, O=system:masters (signed by kubernetes-ca-client)
	APIServerKubeletClientCert, APIServerKubeletClientKey string

	// Keypair used to verify authenticity of cluster messages (e.g. for configmap/k8sd-config)
	K8sdPublicKey, K8sdPrivateKey string
}

type ControlPlanePKIOpts struct {
	Hostname                  string
	DNSSANs                   []string
	IPSANs                    []net.IP
	Years                     int
	AllowSelfSignedCA         bool
	IncludeMachineAddressSANs bool
}

func NewControlPlanePKI(opts ControlPlanePKIOpts) *ControlPlanePKI {
	if opts.Years == 0 {
		opts.Years = 1
	}

	return &ControlPlanePKI{
		hostname:                  opts.Hostname,
		years:                     opts.Years,
		ipSANs:                    opts.IPSANs,
		dnsSANs:                   opts.DNSSANs,
		allowSelfSignedCA:         opts.AllowSelfSignedCA,
		includeMachineAddressSANs: opts.IncludeMachineAddressSANs,
	}
}

// CompleteCertificates generates missing or unset certificates. If only a certificate is set and not a key, we assume that the cluster is using managed certificates.
func (c *ControlPlanePKI) CompleteCertificates() error {
	// Fail hard if keys of self-signed certificates are set without the respective certificates
	switch {
	case c.CACert == "" && c.CAKey != "":
		return fmt.Errorf("kubernetes CA key is set without a certificate, fail to prevent causing issues")
	case c.ClientCACert == "" && c.ClientCAKey != "":
		return fmt.Errorf("kubernetes CA client key is set without a certificate, fail to prevent causing issues")
	case c.FrontProxyCACert == "" && c.FrontProxyCAKey != "":
		return fmt.Errorf("front-proxy CA key is set without a certificate, fail to prevent causing issues")
	}

	var machineIPs []net.IP
	if c.includeMachineAddressSANs {
		addresses, err := net.InterfaceAddrs()
		if err != nil {
			return fmt.Errorf("failed to retrieve machine addresses: %w", err)
		}
		for _, addr := range addresses {
			if ip, _, err := net.ParseCIDR(addr.String()); err == nil && ip != nil {
				machineIPs = append(machineIPs, ip)
			}
		}
	} else {
		machineIPs = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
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

	// Generate self-signed client CA (if not set already)
	if c.ClientCACert == "" && c.ClientCAKey == "" {
		if !c.allowSelfSignedCA {
			return fmt.Errorf("kubernetes client CA not specified and generating self-signed CA not allowed")
		}
		cert, key, err := generateSelfSignedCA(pkix.Name{CommonName: "kubernetes-ca-client"}, c.years, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate kubernetes client CA: %w", err)
		}
		c.ClientCACert = cert
		c.ClientCAKey = key
	}

	serverCACert, serverCAKey, err := loadCertificate(c.CACert, c.CAKey)
	if err != nil {
		return fmt.Errorf("failed to parse kubernetes CA: %w", err)
	}

	clientCACert, clientCAKey, err := loadCertificate(c.ClientCACert, c.ClientCAKey)
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

	// Generate service account key (if missing)
	if c.ServiceAccountKey == "" {
		if !c.allowSelfSignedCA {
			return fmt.Errorf("service account signing key not specified and generating new key is not allowed")
		}

		key, _, err := generateRSAKey(2048)
		if err != nil {
			return fmt.Errorf("failed to generate service account key: %w", err)
		}

		c.ServiceAccountKey = key
	}

	// Generate kubelet certificate (if missing)
	if c.KubeletCert == "" || c.KubeletKey == "" {
		if serverCAKey == nil {
			return fmt.Errorf("using an external kubernetes CA without providing the kubelet certificate is not possible")
		}

		template, err := generateCertificate(
			pkix.Name{CommonName: fmt.Sprintf("system:node:%s", c.hostname), Organization: []string{"system:nodes"}},
			c.years, false, append(c.dnsSANs, c.hostname), append(c.ipSANs, machineIPs...),
		)
		if err != nil {
			return fmt.Errorf("failed to generate kubelet certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, serverCACert, &serverCAKey.PublicKey, serverCAKey)
		if err != nil {
			return fmt.Errorf("failed to sign kubelet certificate: %w", err)
		}

		c.KubeletCert = cert
		c.KubeletKey = key
	}

	// Generate apiserver-kubelet-client certificate (if missing)
	if c.APIServerKubeletClientCert == "" || c.APIServerKubeletClientKey == "" {
		if clientCAKey == nil {
			return fmt.Errorf("using an external kubernetes CA without providing the apiserver-kubelet-client certificate is not possible")
		}

		template, err := generateCertificate(pkix.Name{CommonName: "apiserver-kubelet-client", Organization: []string{"system:masters"}}, c.years, false, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to generate apiserver-kubelet-client certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, clientCACert, &clientCAKey.PublicKey, clientCAKey)
		if err != nil {
			return fmt.Errorf("failed to sign apiserver-kubelet-client certificate: %w", err)
		}

		c.APIServerKubeletClientCert = cert
		c.APIServerKubeletClientKey = key
	}

	// Generate kube-apiserver certificate (if missing)
	if c.APIServerCert == "" || c.APIServerKey == "" {
		if serverCAKey == nil {
			return fmt.Errorf("using an external kubernetes CA without providing the apiserver certificate is not possible")
		}

		template, err := generateCertificate(
			pkix.Name{CommonName: "kube-apiserver"},
			c.years,
			false,
			append(c.dnsSANs, "kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster", "kubernetes.default.svc.cluster.local"), append(c.ipSANs, machineIPs...))
		if err != nil {
			return fmt.Errorf("failed to generate apiserver certificate: %w", err)
		}
		cert, key, err := signCertificate(template, 2048, serverCACert, &serverCAKey.PublicKey, serverCAKey)
		if err != nil {
			return fmt.Errorf("failed to sign apiserver certificate: %w", err)
		}

		c.APIServerCert = cert
		c.APIServerKey = key
	}

	for _, i := range []struct {
		name string
		cn   string
		o    []string
		cert *string
		key  *string
	}{
		{name: "admin", cn: "kubernetes-admin", o: []string{"system:masters"}, cert: &c.AdminClientCert, key: &c.AdminClientKey},
		{name: "controller", cn: "system:kube-controller-manager", cert: &c.KubeControllerManagerClientCert, key: &c.KubeControllerManagerClientKey},
		{name: "proxy", cn: "system:kube-proxy", cert: &c.KubeProxyClientCert, key: &c.KubeProxyClientKey},
		{name: "scheduler", cn: "system:kube-scheduler", cert: &c.KubeSchedulerClientCert, key: &c.KubeSchedulerClientKey},
		{name: "kubelet", cn: fmt.Sprintf("system:node:%s", c.hostname), o: []string{"system:nodes"}, cert: &c.KubeletClientCert, key: &c.KubeletClientKey},
	} {
		if *i.cert == "" || *i.key == "" {
			if clientCAKey == nil {
				return fmt.Errorf("using an external kubernetes CA client without providing the %s certificate is not possible", i.name)
			}

			template, err := generateCertificate(pkix.Name{CommonName: i.cn, Organization: i.o}, c.years, false, nil, nil)
			if err != nil {
				return fmt.Errorf("failed to generate %s client certificate: %w", i.name, err)
			}

			cert, key, err := signCertificate(template, 2048, clientCACert, &clientCAKey.PublicKey, clientCAKey)
			if err != nil {
				return fmt.Errorf("failed to sign %s client certificate: %w", i.name, err)
			}

			*i.cert = cert
			*i.key = key
		}
	}

	// Generate k8sd cluster key-pair (if missing)
	if c.K8sdPrivateKey == "" || c.K8sdPublicKey == "" {
		if !c.allowSelfSignedCA {
			return fmt.Errorf("cluster keypair not specified and generating new key is not allowed")
		}

		priv, pub, err := generateRSAKey(2048)
		if err != nil {
			return fmt.Errorf("failed to generate cluster keypair: %w", err)
		}

		c.K8sdPrivateKey = priv
		c.K8sdPublicKey = pub
	}

	return nil
}
