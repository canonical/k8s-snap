package component

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/vals"
)

func UpdateNetworkComponent(ctx context.Context, s snap.Snap, isRefresh bool, podCIDR string) error {
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
	} else {
		p, err := utils.GetMountPropagation("/sys")
		if err != nil {
			return fmt.Errorf("failed to get mount propagation for %s: %w", p, err)
		}
		if p == "private" {
			onLXD, err := s.OnLXD(ctx)
			if err != nil {
				log.Printf("failed to check if on lxd: %v", err)
			}
			if onLXD {
				return fmt.Errorf("/sys is not a shared mount on the LXD container, this might be resolved by updating LXD on the host to version 5.0.2 or newer")
			}
			return fmt.Errorf("/sys is not a shared mount")
		}
	}

	if isRefresh {
		if err := manager.Refresh("network", values); err != nil {
			return fmt.Errorf("failed to refresh network component: %w", err)
		}
	} else {
		if err := manager.Enable("network", values); err != nil {
			return fmt.Errorf("failed to enable network component: %w", err)
		}
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

func ReconcileNetworkComponent(ctx context.Context, s snap.Snap, alreadyEnabled *bool, requestEnabled *bool, clusterConfig types.ClusterConfig) error {
	if vals.OptionalBool(requestEnabled, true) && vals.OptionalBool(alreadyEnabled, false) {
		// If already enabled, and request does not contain `enabled` key
		// or if already enabled and request contains `enabled=true`
		err := UpdateNetworkComponent(ctx, s, true, clusterConfig.Network.PodCIDR)
		if err != nil {
			return fmt.Errorf("failed to refresh network: %w", err)
		}
		return nil
	} else if vals.OptionalBool(requestEnabled, false) {
		err := UpdateNetworkComponent(ctx, s, false, clusterConfig.Network.PodCIDR)
		if err != nil {
			return fmt.Errorf("failed to enable network: %w", err)
		}
		return nil
	} else if !vals.OptionalBool(requestEnabled, false) {
		err := DisableNetworkComponent(s)
		if err != nil {
			return fmt.Errorf("failed to disable network: %w", err)
		}
		return nil
	}
	return nil
}
