package pki

import (
	"crypto/x509/pkix"
	"fmt"
	"net"
)

type WorkerNodePKI struct {
	CACert string // CN=kubernetes-ca

	// [server] CN=system:node:hostname, O=system:nodes, DNS=hostname, IP=127.0.0.1,address (signed by kubernetes-ca)
	KubeletCert, KubeletKey string

	// [client] CN=system:kube-proxy (signed by kubernetes-ca-client)
	KubeProxyClientCert, KubeProxyClientKey string

	// [client] CN=system:node:hostname, O=system:nodes (signed by kubernetes-ca-client)
	KubeletClientCert, KubeletClientKey string
}

// CompleteWorkerNodePKI generates the PKI needed for a worker node.
func (c *ControlPlanePKI) CompleteWorkerNodePKI(hostname string, nodeIP net.IP, bits int) (*WorkerNodePKI, error) {
	serverCACert, serverCAKey, err := loadCertificate(c.CACert, c.CAKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes CA: %w", err)
	}

	clientCACert, clientCAKey, err := loadCertificate(c.ClientCACert, c.ClientCAKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes client CA: %w", err)
	}

	pki := &WorkerNodePKI{CACert: c.CACert}

	// we have a cluster CA key, sign the kubelet server certificate
	if serverCAKey != nil {
		template, err := generateCertificate(pkix.Name{CommonName: fmt.Sprintf("system:node:%s", hostname), Organization: []string{"system:nodes"}}, c.years, false, []string{hostname}, []net.IP{{127, 0, 0, 1}, nodeIP})
		if err != nil {
			return nil, fmt.Errorf("failed to generate kubelet certificate for hostname=%s address=%s: %w", hostname, nodeIP.String(), err)
		}
		cert, key, err := signCertificate(template, bits, serverCACert, &serverCAKey.PublicKey, serverCAKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign kubelet certificate for hostname=%s address=%s: %w", hostname, nodeIP.String(), err)
		}
		pki.KubeletCert = cert
		pki.KubeletKey = key
	}

	// we have a client CA key, sign the kubelet and kube-proxy client certificates
	if clientCAKey != nil {
		for _, i := range []struct {
			name string
			cn   string
			o    []string
			cert *string
			key  *string
		}{
			{name: "proxy", cn: "system:kube-proxy", cert: &c.KubeProxyClientCert, key: &c.KubeProxyClientKey},
			{name: "kubelet", cn: fmt.Sprintf("system:node:%s", c.hostname), o: []string{"system:nodes"}, cert: &c.KubeletClientCert, key: &c.KubeletClientKey},
		} {
			if *i.cert == "" || *i.key == "" {
				template, err := generateCertificate(pkix.Name{CommonName: i.cn, Organization: i.o}, c.years, false, nil, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to generate %s client certificate: %w", i.name, err)
				}

				cert, key, err := signCertificate(template, 2048, clientCACert, &clientCAKey.PublicKey, clientCAKey)
				if err != nil {
					return nil, fmt.Errorf("failed to sign %s client certificate: %w", i.name, err)
				}

				*i.cert = cert
				*i.key = key
			}
		}
	}

	return pki, nil
}

func (c *WorkerNodePKI) CompleteCertificates() error {
	if c.CACert == "" {
		return fmt.Errorf("kubernetes CA not specified")
	}
	if c.KubeletCert == "" || c.KubeletKey == "" {
		return fmt.Errorf("kubelet certificate not specified")
	}
	if c.KubeletClientCert == "" || c.KubeletClientKey == "" {
		return fmt.Errorf("kubelet client certificate not specified")
	}
	if c.KubeProxyClientCert == "" || c.KubeProxyClientKey == "" {
		return fmt.Errorf("kube-proxy client certificate not specified")
	}
	return nil
}
