package pki

import (
	"crypto/x509/pkix"
	"fmt"
	"net"
)

type WorkerNodePKI struct {
	CACert      string
	KubeletCert string
	KubeletKey  string
}

// CompleteWorkerNodePKI generates the PKI needed for a worker node.
func (c *ControlPlanePKI) CompleteWorkerNodePKI(hostname string, nodeIP net.IP, bits int) (*WorkerNodePKI, error) {
	caCert, caKey, err := loadCertificate(c.CACert, c.CAKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes CA: %w", err)
	}

	// we do not have a CA key to sign the kubelet certificate, only send the cluster CA
	if caKey == nil {
		return &WorkerNodePKI{CACert: c.CACert}, nil
	}

	template, err := generateCertificate(pkix.Name{CommonName: fmt.Sprintf("system:node:%s", hostname), Organization: []string{"system:nodes"}}, c.years, false, []string{hostname}, []net.IP{{127, 0, 0, 1}, nodeIP})
	if err != nil {
		return nil, fmt.Errorf("failed to generate kubelet certificate for hostname=%s address=%s: %w", hostname, nodeIP.String(), err)
	}
	cert, key, err := signCertificate(template, bits, caCert, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign kubelet certificate for hostname=%s address=%s: %w", hostname, nodeIP.String(), err)
	}

	return &WorkerNodePKI{
		CACert:      c.CACert,
		KubeletCert: cert,
		KubeletKey:  key,
	}, nil
}

func (c *WorkerNodePKI) CompleteCertificates() error {
	if c.CACert == "" {
		return fmt.Errorf("kubernetes CA not specified")
	}
	if c.KubeletCert == "" || c.KubeletKey == "" {
		return fmt.Errorf("kubelet certificate not specified")
	}
	return nil
}

func (c *WorkerNodePKI) IsKubeletPresent() bool {
	return c.KubeletCert != "" && c.KubeletKey != ""
}
