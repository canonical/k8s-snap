package utils

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/canonical/k8s/pkg/snap"
)

// GenerateX509Kubeconfig creates a kubeconfig file with the given x509 certificate and key data.
func GenerateX509Kubeconfig(keyPem, certPem, caCertPem []byte, path string) error {
	val, err := GetServiceArgument("kube-apiserver", "--secure-port")
	if err != nil {
		return fmt.Errorf("failed while getting apiserver port: %w", err)
	}

	port, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("apiserver port is not an integer: %w", err)
	}

	return TemplateAndSave(snap.Path("k8s/config/kubeconfig-with-x509.tmpl"),
		struct {
			CaData        string
			ApiServerIp   string
			ApiServerPort int
			CertData      string
			KeyData       string
		}{
			CaData:        base64.StdEncoding.EncodeToString(caCertPem),
			ApiServerIp:   "127.0.0.1",
			ApiServerPort: port,
			CertData:      base64.StdEncoding.EncodeToString(certPem),
			KeyData:       base64.StdEncoding.EncodeToString(keyPem),
		},
		path,
	)
}

// GenerateKubeconfig creates a kubeconfig file with the given token and CA data.
func GenerateKubeconfig(token string, caCertPem []byte, path string) error {
	val, err := GetServiceArgument("kube-apiserver", "--secure-port")
	if err != nil {
		return fmt.Errorf("failed while getting apiserver port: %w", err)
	}

	port, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("apiserver port is not an integer: %w", err)
	}

	return TemplateAndSave(snap.Path("k8s/config/kubeconfig-with-token.tmpl"),
		struct {
			CaData        string
			ApiServerIp   string
			ApiServerPort int
			Token         string
		}{
			CaData:        base64.StdEncoding.EncodeToString(caCertPem),
			ApiServerIp:   "127.0.0.1",
			ApiServerPort: port,
			Token:         token,
		},
		path,
	)
}
