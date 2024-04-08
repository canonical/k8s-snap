package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/k8s/pkg/utils/vals"
)

// UpdateDNSComponent enables or refreshes DNS on the cluster.
// On success, it returns the IP of the DNS service and the cluster domain.
func UpdateDNSComponent(ctx context.Context, s snap.Snap, isRefresh bool, clusterDomain, serviceIP string, upstreamNameservers []string) (string, string, error) {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to get component manager: %w", err)
	}

	upstreamNameserver := "/etc/resolv.conf"
	if clusterDomain == "" {
		clusterDomain = "cluster.local"
	}

	if len(upstreamNameservers) > 0 {
		upstreamNameserver = strings.Join(upstreamNameservers, " ")
	}

	values := map[string]any{
		"image": map[string]any{
			"repository": dnsImageRepository,
			"tag":        dnsImageTag,
		},
		"service": map[string]any{
			"name": "coredns",
		},
		"deployment": map[string]any{
			"name": "coredns",
		},
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
						"parameters":  fmt.Sprintf("%s in-addr.arpa ip6.arpa", clusterDomain),
						"configBlock": "pods insecure\nfallthrough in-addr.arpa ip6.arpa\nttl 30",
					},
					{"name": "prometheus", "parameters": "0.0.0.0:9153"},
					{"name": "forward", "parameters": fmt.Sprintf(". %s", upstreamNameserver)},
					{"name": "cache", "parameters": "30"},
					{"name": "loop"},
					{"name": "reload"},
					{"name": "loadbalance"},
				},
			},
		},
	}

	if serviceIP != "" {
		service := values["service"].(map[string]any)
		service["clusterIP"] = serviceIP
	}

	if isRefresh {
		if err := manager.Refresh("dns", values); err != nil {
			return "", "", fmt.Errorf("failed to refresh dns component: %w", err)
		}
	} else {
		if err := manager.Enable("dns", values); err != nil {
			return "", "", fmt.Errorf("failed to enable dns component: %w", err)
		}
	}

	client, err := k8s.NewClient(s.KubernetesRESTClientGetter(""))
	if err != nil {
		return "", "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	dnsIP, err := client.GetServiceClusterIP(ctx, "coredns", "kube-system")
	if err != nil {
		return "", "", fmt.Errorf("failed to get dns service: %w", err)
	}

	// TODO: this should propagate to all nodes
	changed, err := snaputil.UpdateServiceArguments(s, "kubelet", map[string]string{
		"--cluster-dns":    dnsIP,
		"--cluster-domain": clusterDomain,
	}, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to update kubelet arguments: %w", err)
	}

	if changed {
		if err := s.RestartService(ctx, "kubelet"); err != nil {
			return "", "", fmt.Errorf("failed to restart kubelet to apply new dns configuration: %w", err)
		}
	}
	return dnsIP, clusterDomain, nil
}

func DisableDNSComponent(ctx context.Context, s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	if err := manager.Disable("dns"); err != nil {
		return fmt.Errorf("failed to disable dns component: %w", err)
	}

	changed, err := snaputil.UpdateServiceArguments(s, "kubelet", map[string]string{"--cluster-domain": "cluster.local"}, []string{"--cluster-dns"})
	if err != nil {
		return fmt.Errorf("failed to update kubelet arguments: %w", err)
	}

	if changed {
		if err := s.RestartService(ctx, "kubelet"); err != nil {
			return fmt.Errorf("failed to restart service 'kubelet': %w", err)
		}
	}

	return nil
}

func ReconcileDNSComponent(ctx context.Context, s snap.Snap, alreadyEnabled *bool, requestEnabled *bool, clusterConfig types.ClusterConfig) (string, string, error) {
	if vals.OptionalBool(requestEnabled, true) && vals.OptionalBool(alreadyEnabled, false) {
		// If already enabled, and request does not contain `enabled` key
		// or if already enabled and request contains `enabled=true`
		dnsIP, clusterDomain, err := UpdateDNSComponent(ctx, s, true, clusterConfig.Kubelet.GetClusterDomain(), clusterConfig.Kubelet.GetClusterDNS(), clusterConfig.DNS.GetUpstreamNameservers())
		if err != nil {
			return "", "", fmt.Errorf("failed to refresh dns: %w", err)
		}
		return dnsIP, clusterDomain, nil
	} else if vals.OptionalBool(requestEnabled, false) {
		dnsIP, clusterDomain, err := UpdateDNSComponent(ctx, s, false, clusterConfig.Kubelet.GetClusterDomain(), clusterConfig.Kubelet.GetClusterDNS(), clusterConfig.DNS.GetUpstreamNameservers())
		if err != nil {
			return "", "", fmt.Errorf("failed to enable dns: %w", err)
		}
		return dnsIP, clusterDomain, nil
	} else if !vals.OptionalBool(requestEnabled, false) {
		err := DisableDNSComponent(ctx, s)
		if err != nil {
			return "", "", fmt.Errorf("failed to disable dns: %w", err)
		}
		return clusterConfig.Kubelet.GetClusterDNS(), clusterConfig.Kubelet.GetClusterDomain(), nil
	}
	return "", "", nil
}
