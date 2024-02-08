package setup

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
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

// writeKubeconfigToFile writes a kubeconfig file to disk.
func writeKubeconfigToFile(path string, token string, url string, caPEM string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	return renderKubeconfig(file, token, url, caPEM)
}
