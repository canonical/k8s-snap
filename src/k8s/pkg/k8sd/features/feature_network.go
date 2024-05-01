package features

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
)

// ApplyNetwork is used to configure the CNI feature on Canonical Kubernetes.
// ApplyNetwork will deploy Cilium when cfg.Enabled is true.
// ApplyNetwork will remove Cilium when cfg.Enabled is false.
// ApplyNetwork requires that bpf and cgroups2 are already mounted and available when running under strict snap confinement. If they are not, it will fail (since Cilium will not have the required permissions to mount them).
// ApplyNetwork requires that `/sys` is mounted as a shared mount when running under classic snap confinement. This is to ensure that Cilium will be able to automatically mount bpf and cgroups2 on the pods.
// ApplyNetwork returns an error if anything fails.
func ApplyNetwork(ctx context.Context, snap snap.Snap, cfg types.Network) error {
	m := snap.HelmClient()

	if !cfg.GetEnabled() {
		if _, err := m.Apply(ctx, chartCilium, helm.StateDeleted, nil); err != nil {
			return fmt.Errorf("failed to uninstall network: %w", err)
		}
		return nil
	}

	clusterCIDRs := strings.Split(cfg.GetPodCIDR(), ",")
	if v := len(clusterCIDRs); v != 1 && v != 2 {
		return fmt.Errorf("invalid kube-proxy --cluster-cidr value: %v", clusterCIDRs)
	}

	var (
		ipv4CIDR string
		ipv6CIDR string
	)
	for _, cidr := range clusterCIDRs {
		_, parsed, err := net.ParseCIDR(cidr)
		switch {
		case err != nil:
			return fmt.Errorf("failed to parse cidr: %w", err)
		case parsed.IP.To4() != nil:
			ipv4CIDR = cidr
		default:
			ipv6CIDR = cidr
		}
	}

	values := map[string]any{
		"image": map[string]any{
			"repository": ciliumAgentImageRepository,
			"tag":        ciliumAgentImageTag,
			"useDigest":  false,
		},
		"socketLB": map[string]any{
			"enabled": true,
		},
		"cni": map[string]any{
			"confPath": "/etc/cni/net.d",
			"binPath":  "/opt/cni/bin",
		},
		"operator": map[string]any{
			"replicas": 1,
			"image": map[string]any{
				"repository": ciliumOperatorImageRepository,
				"tag":        ciliumOperatorImageTag,
				"useDigest":  false,
			},
		},
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": ipv4CIDR,
				"clusterPoolIPv6PodCIDRList": ipv6CIDR,
			},
		},
		"nodePort": map[string]any{
			"enabled": true,
		},
		"disableEnvoyVersionCheck": true,
	}

	if snap.Strict() {
		bpfMnt, err := utils.GetMountPath("bpf")
		if err != nil {
			return fmt.Errorf("failed to get bpf mount path: %w", err)
		}

		cgrMnt, err := utils.GetMountPath("cgroup2")
		if err != nil {
			return fmt.Errorf("failed to get cgroup2 mount path: %w", err)
		}

		values["bpf"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"root": bpfMnt,
		}
		values["cgroup"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"hostRoot": cgrMnt,
		}
	} else {
		p, err := utils.GetMountPropagation("/sys")
		if err != nil {
			return fmt.Errorf("failed to get mount propagation for %s: %w", p, err)
		}
		if p == "private" {
			onLXD, err := snap.OnLXD(ctx)
			if err != nil {
				log.Printf("failed to check if on LXD: %v", err)
			}
			if onLXD {
				return fmt.Errorf("/sys is not a shared mount on the LXD container, this might be resolved by updating LXD on the host to version 5.0.2 or newer")
			}
			return fmt.Errorf("/sys is not a shared mount")
		}
	}

	if _, err := m.Apply(ctx, chartCilium, helm.StatePresent, values); err != nil {
		return fmt.Errorf("failed to enable network: %w", err)
	}

	return nil
}

func rolloutRestartCilium(ctx context.Context, snap snap.Snap, attempts int) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDeployment(ctx, "cilium-operator", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart cilium-operator deployment after %d attempts: %w", attempts, err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDaemonset(ctx, "cilium", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium daemonset: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart cilium daemonset after %d attempts: %w", attempts, err)
	}

	return nil
}
