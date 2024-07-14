package setup

import (
	"context"
	"fmt"
	"path"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
)

// KubeProxy configures kube-proxy on the local node.
func KubeProxy(ctx context.Context, snap snap.Snap, hostname string, podCIDR string, extraArgs map[string]*string) error {
	serviceArgs := map[string]string{
		"--cluster-cidr":         podCIDR,
		"--healthz-bind-address": "127.0.0.1",
		"--kubeconfig":           path.Join(snap.KubernetesConfigDir(), "proxy.conf"),
		"--profiling":            "false",
	}

	if hostname != snap.Hostname() {
		serviceArgs["--hostname-override"] = hostname
	}
	onLXD, err := snap.OnLXD(ctx)
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to check if running on LXD")
	}
	if onLXD {
		// A container cannot set this sysctl config in LXD. So, we disable it by setting it to "0".
		// See: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-proxy/
		serviceArgs["--conntrack-max-per-core"] = "0"
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "kube-proxy", serviceArgs, nil); err != nil {
		return fmt.Errorf("failed to render arguments file: %w", err)
	}

	// Apply extra arguments after the defaults, so they can override them.
	updateArgs, deleteArgs := utils.ServiceArgsFromMap(extraArgs)
	if _, err := snaputil.UpdateServiceArguments(snap, "kube-proxy", updateArgs, deleteArgs); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}
	return nil
}
