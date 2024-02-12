package setup

import (
	"fmt"
	"path"

	"github.com/canonical/k8s/pkg/proxy"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

// K8sAPIServerProxy prepares configuration for k8s-apiserver-proxy.
func K8sAPIServerProxy(snap snap.Snap, servers []string) error {
	configFile := path.Join(snap.ServiceExtraConfigDir(), "k8s-apiserver-proxy.json")
	if err := proxy.WriteEndpointsConfig(servers, configFile); err != nil {
		return fmt.Errorf("failed to write proxy configuration file: %w", err)
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "k8s-apiserver-proxy", map[string]string{
		"--listen":     "127.0.0.1:6443",
		"--kubeconfig": path.Join(snap.KubernetesConfigDir(), "kubelet.conf"),
		"--endpoints":  configFile,
	}, nil); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}

	return nil
}
