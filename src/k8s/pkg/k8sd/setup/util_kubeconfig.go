package setup

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/pki"
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

// SetupControlPlaneKubeconfigs writes kubeconfig files for the control plane components.
func SetupControlPlaneKubeconfigs(kubeConfigDir string, localhostAddress string, securePort int, pki pki.ControlPlanePKI) error {
	for _, kubeconfig := range []struct {
		file string
		crt  string
		key  string
	}{
		{file: "admin.conf", crt: pki.AdminClientCert, key: pki.AdminClientKey},
		{file: "controller.conf", crt: pki.KubeControllerManagerClientCert, key: pki.KubeControllerManagerClientKey},
		{file: "proxy.conf", crt: pki.KubeProxyClientCert, key: pki.KubeProxyClientKey},
		{file: "scheduler.conf", crt: pki.KubeSchedulerClientCert, key: pki.KubeSchedulerClientKey},
		{file: "kubelet.conf", crt: pki.KubeletClientCert, key: pki.KubeletClientKey},
	} {
		if err := Kubeconfig(filepath.Join(kubeConfigDir, kubeconfig.file), fmt.Sprintf("%s:%d", localhostAddress, securePort), pki.CACert, kubeconfig.crt, kubeconfig.key); err != nil {
			return fmt.Errorf("failed to write kubeconfig %s: %w", kubeconfig.file, err)
		}
	}
	return nil
}
