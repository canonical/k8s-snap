package setup

import (
	"context"
	"fmt"
	"path"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

// KubeProxy configures kube-proxy on the local node.
func KubeProxy(ctx context.Context, snap snap.Snap, hostname string, podCIDR string) error {
	serviceArgs := map[string]string{
		"--cluster-cidr":         podCIDR,
		"--healthz-bind-address": "127.0.0.1",
		"--hostname-override":    hostname,
		"--kubeconfig":           path.Join(snap.KubernetesConfigDir(), "proxy.conf"),
		"--profiling":            "false",
	}

	onLXD, err := snap.OnLXD(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if on lxd: %w", err)
	}
	if onLXD {
		// A container cannot set this sysctl config in LXD. So, we disable it by setting it to "0".
		// See: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-proxy/
		serviceArgs["--conntrack-max-per-core"] = "0"
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "kube-proxy", serviceArgs, nil); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}
	return nil
}
