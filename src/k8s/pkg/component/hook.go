package component

import (
	"fmt"
	"net"
	"strings"

	"github.com/canonical/k8s/pkg/utils"
)

type valuesHook func() (map[string]any, error)

var valuesHooks = map[string]valuesHook{
	"network": networkValues,
	"dns":     dnsValues,
}

func dnsValues() (map[string]any, error) {
	values := map[string]any{
		"servers": []map[string]any{
			{
				"zones": []map[string]any{
					{"zone": "."},
				},
				"port": 53,
				"plugins": []map[string]any{
					{"name": "errors"},
					{"name": "health", "configBlock": "lameduck 5s"},
					{"name": "ready"},
					{
						"name":        "kubernetes",
						"parameters":  "cluster.local in-addr.arpa ip6.arpa",
						"configBlock": "pods insecure\nfallthrough in-addr.arpa ip6.arpa\nttl 30",
					},
					{"name": "prometheus", "parameters": "0.0.0.0:9153"},
					{"name": "forward", "parameters": fmt.Sprintf(". %s", "/etc/resolv.conf")},
					{"name": "cache", "parameters": "30"},
					{"name": "loop"},
					{"name": "reload"},
					{"name": "loadbalance"},
				},
			},
		},
	}
	return values, nil
}

func networkValues() (map[string]any, error) {
	bpfMnt, err := utils.GetMountPath("bpf")
	if err != nil {
		return nil, fmt.Errorf("failed to get bpf mount path: %w", err)
	}

	cgrMnt, err := utils.GetMountPath("cgroup2")
	if err != nil {
		return nil, fmt.Errorf("failed to get cgroup2 mount path: %w", err)
	}

	// TODO: the cluster cidr should be configurable through a common interface
	clusterCIDRStr, err := utils.GetServiceArgument("kube-proxy", "--cluster-cidr")
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster cidrs from kube-proxy arguments: %w", err)
	}

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
			"runPath": utils.SnapCommonPath("var", "run", "cilium"),
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
