package component

import (
	"fmt"
	"net"
	"strings"

	s "github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

type valuesHook func(s.Snap) (map[string]any, error)

var valuesHooks = map[string]valuesHook{
	"network": networkValues,
}

func networkValues(snap s.Snap) (map[string]any, error) {
	bpfMnt, err := utils.GetMountPath("bpf")
	if err != nil {
		return nil, fmt.Errorf("failed to get bpf mount path: %w", err)
	}

	cgrMnt, err := utils.GetMountPath("cgroup2")
	if err != nil {
		return nil, fmt.Errorf("failed to get cgroup2 mount path: %w", err)
	}

	// TODO: the cluster cidr should be configurable through a common interface
	clusterCIDRStr := s.GetServiceArgument(snap, "kube-proxy", "--cluster-cidr")
	clusterCIDRs := strings.Split(clusterCIDRStr, ",")
	if v := len(clusterCIDRs); v != 1 && v != 2 {
		return nil, fmt.Errorf("invalid kube-proxy --cluster-cidr value: %v", clusterCIDRs)
	}

	var (
		ipv4CIDR string
		ipv6CIDR string
	)
	for _, cidr := range clusterCIDRs {
		_, parsed, err := net.ParseCIDR(cidr)
		switch {
		case err != nil:
			return nil, fmt.Errorf("failed to parse cidr: %w", err)
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
			"runPath": snap.CommonPath("var", "run", "cilium"),
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

	return values, nil
}
