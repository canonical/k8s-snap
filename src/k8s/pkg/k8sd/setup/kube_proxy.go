package setup

import (
	"fmt"
	"path"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

// KubeProxy configures kube-proxy on the local node.
func KubeProxy(snap snap.Snap, caPEM string, hostname string, token string, clusterCIDR string) error {
	if err := writeKubeconfigToFile(path.Join(snap.KubernetesConfigDir(), "proxy.conf"), token, "127.0.0.1:6443", caPEM); err != nil {
		return fmt.Errorf("failed to write proxy.conf: %w", err)
	}
	if _, err := snaputil.UpdateServiceArguments(snap, "kube-proxy", map[string]string{
		"--kubeconfig":           path.Join(snap.KubernetesConfigDir(), "proxy.conf"),
		"--cluster-cidr":         clusterCIDR,
		"--healthz-bind-address": "127.0.0.1",
		"--profiling":            "false",
		"--hostname-override":    hostname,
	}, nil); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}
	return nil
}
