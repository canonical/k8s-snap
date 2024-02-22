package setup

import (
	"fmt"
	"path"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

// KubeProxy configures kube-proxy on the local node.
func KubeProxy(snap snap.Snap, hostname string, podCIDR string) error {
	if _, err := snaputil.UpdateServiceArguments(snap, "kube-proxy", map[string]string{
		"--cluster-cidr":         podCIDR,
		"--healthz-bind-address": "127.0.0.1",
		"--hostname-override":    hostname,
		"--kubeconfig":           path.Join(snap.KubernetesConfigDir(), "proxy.conf"),
		"--profiling":            "false",
	}, nil); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}
	return nil
}
