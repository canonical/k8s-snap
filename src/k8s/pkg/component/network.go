package component

import (
	"fmt"
	"net"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

func EnableNetworkComponent(s snap.Snap, podCIDR string) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	clusterCIDRs := strings.Split(podCIDR, ",")
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
	}

	if s.Strict() {
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
	}

	if err := manager.Enable("network", values); err != nil {
		return fmt.Errorf("failed to enable network component: %w", err)
	}

	return nil
}

func DisableNetworkComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	if err := manager.Disable("network"); err != nil {
		return fmt.Errorf("failed to disable network component: %w", err)
	}

	return nil
}
