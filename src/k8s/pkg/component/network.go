package component

import (
	"fmt"
	"net"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

func EnableNetworkComponent(s snap.Snap) error {
	manager, err := NewManager(s)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	bpfMnt, err := utils.GetMountPath("bpf")
	if err != nil {
		return fmt.Errorf("failed to get bpf mount path: %w", err)
	}

	cgrMnt, err := utils.GetMountPath("cgroup2")
	if err != nil {
		return fmt.Errorf("failed to get cgroup2 mount path: %w", err)
	}

	// TODO: the cluster cidr should be configurable through a common interface
	clusterCIDRStr := snap.GetServiceArgument(s, "kube-proxy", "--cluster-cidr")
	clusterCIDRs := strings.Split(clusterCIDRStr, ",")
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
		"cni": map[string]any{
			"confPath": "/etc/cni/net.d",
			"binPath":  "/opt/cni/bin",
		},
		"daemon": map[string]any{
			"runPath": s.CommonPath("var", "run", "cilium"),
		},
		"operator": map[string]any{
			"replicas": 1,
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
		"bpf": map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"root": bpfMnt,
		},
		"cgroup": map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"hostRoot": cgrMnt,
		},
		"l2announcements": map[string]any{
			"enabled": true,
		},
	}

	err = manager.Enable("network", values)
	if err != nil {
		return fmt.Errorf("failed to enable network component: %w", err)
	}

	return nil
}

func DisableNetworkComponent(s snap.Snap) error {
	manager, err := NewManager(s)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	err = manager.Disable("network")
	if err != nil {
		return fmt.Errorf("failed to enable network component: %w", err)
	}

	return nil
}
