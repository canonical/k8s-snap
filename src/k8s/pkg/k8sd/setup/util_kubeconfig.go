package setup

import (
	"fmt"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// createConfig generates a Config suitable for our k8s environment.
func createConfig(token string, server string, caPEM string) *clientcmdapi.Config {
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
		Token: token,
	}
	config.Contexts["k8s"] = &clientcmdapi.Context{
		Cluster:  "k8s",
		AuthInfo: "k8s-user",
	}
	config.CurrentContext = "k8s"

	return config
}

// Kubeconfig writes a kubeconfig file to disk.
func Kubeconfig(path string, token string, url string, caPEM string) error {
	config := createConfig(token, url, caPEM)
	if err := clientcmd.WriteToFile(*config, path); err != nil {
		return fmt.Errorf("failed to write kubeconfig: %w", err)
	}
	return nil
}

// KubeconfigString provides a stringified kubeconfig.
func KubeconfigString(token string, url string, caPEM string) (string, error) {
	config := createConfig(token, url, caPEM)
	kubeconfig, err := clientcmd.Write(*config)
	if err != nil {
		return "", err
	}
	return string(kubeconfig), nil
}
