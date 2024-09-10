package setup

import (
	"fmt"
	"path/filepath"

	"github.com/canonical/k8s/pkg/proxy"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
)

// K8sAPIServerProxy prepares configuration for k8s-apiserver-proxy.
func K8sAPIServerProxy(snap snap.Snap, servers []string, extraArgs map[string]*string) error {
	configFile := filepath.Join(snap.ServiceExtraConfigDir(), "k8s-apiserver-proxy.json")
	if err := proxy.WriteEndpointsConfig(servers, configFile); err != nil {
		return fmt.Errorf("failed to write proxy configuration file: %w", err)
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "k8s-apiserver-proxy", map[string]string{
		"--endpoints":  configFile,
		"--kubeconfig": filepath.Join(snap.KubernetesConfigDir(), "kubelet.conf"),
		"--listen":     "[::1]:6443",
	}, nil); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}

	// Apply extra arguments after the defaults, so they can override them.
	updateArgs, deleteArgs := utils.ServiceArgsFromMap(extraArgs)
	if _, err := snaputil.UpdateServiceArguments(snap, "k8s-apiserver-proxy", updateArgs, deleteArgs); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}
	return nil
}
