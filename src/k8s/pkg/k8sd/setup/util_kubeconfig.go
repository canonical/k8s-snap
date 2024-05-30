package setup

import (
	"fmt"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// createConfig generates a Config suitable for our k8s environment.
func createConfig(server string, caPEM string, crtPEM string, keyPEM string) *clientcmdapi.Config {
	config := clientcmdapi.NewConfig()

	// Default to https:// prefix if no http-like scheme is present.
	// Note: scheme-less host:port isn't a valid url, so no url.Parse here.
	if !strings.HasPrefix(server, "http") {
		server = fmt.Sprintf("https://%s", server)
	}

	config.Clusters["k8s"] = &clientcmdapi.Cluster{
		CertificateAuthorityData: []byte(caPEM),
		Server:                   server,
	}
	config.AuthInfos["k8s-user"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: []byte(crtPEM),
		ClientKeyData:         []byte(keyPEM),
	}
	config.Contexts["k8s"] = &clientcmdapi.Context{
		Cluster:  "k8s",
		AuthInfo: "k8s-user",
	}
	config.CurrentContext = "k8s"

	return config
}

// Kubeconfig writes a kubeconfig file to disk.
func Kubeconfig(path string, url string, caPEM string, crtPEM string, keyPEM string) error {
	config := createConfig(url, caPEM, crtPEM, keyPEM)
	if err := clientcmd.WriteToFile(*config, path); err != nil {
		return fmt.Errorf("failed to write kubeconfig: %w", err)
	}
	return nil
}

// KubeconfigString provides a stringified kubeconfig.
func KubeconfigString(url string, caPEM string, crtPEM string, keyPEM string) (string, error) {
	config := createConfig(url, caPEM, crtPEM, keyPEM)
	kubeconfig, err := clientcmd.Write(*config)
	if err != nil {
		return "", fmt.Errorf("failed to encode kubeconfig yaml: %w", err)
	}
	return string(kubeconfig), nil
}
