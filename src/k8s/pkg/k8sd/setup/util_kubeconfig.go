package setup

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	kubeconfigTemplate = mustTemplate("kubeconfig")
)

type kubeconfigTemplateConfig struct {
	CA    string
	URL   string
	Token string
}

// renderKubeconfig writes a kubeconfig to the specified writer.
func renderKubeconfig(writer io.Writer, token string, url string, caPEM string) error {
	if err := kubeconfigTemplate.Execute(writer, kubeconfigTemplateConfig{
		CA:    base64.StdEncoding.EncodeToString([]byte(caPEM)),
		URL:   url,
		Token: token,
	}); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}
	return nil
}

// createConfig generates a Config suitable for our k8s environment.
func createConfig(token string, url string, caPEM string) *clientcmdapi.Config {
	config := clientcmdapi.NewConfig()
	config.Clusters["k8s"] = &clientcmdapi.Cluster{
		CertificateAuthorityData: []byte(caPEM),
		Server:                   url,
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
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	return renderKubeconfig(file, token, url, caPEM)
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
