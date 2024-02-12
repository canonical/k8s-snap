package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
)

func EnableDNSComponent(ctx context.Context, s snap.Snap, clusterDomain, serviceIP string, upstreamNameservers []string) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	upstreamNameserver := "/etc/resolv.conf"
	if clusterDomain == "" {
		clusterDomain = "cluster.local"
	}

	if len(upstreamNameservers) > 0 {
		upstreamNameserver = strings.Join(upstreamNameservers, " ")
	}

	values := map[string]any{
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

	err = manager.Enable("dns", values)
	if err != nil {
		return fmt.Errorf("failed to enable dns component: %w", err)
	}

	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	dnsIP, err := k8s.GetServiceClusterIP(ctx, client, "coredns", "kube-system")
	if err != nil {
		return fmt.Errorf("failed to get dns service: %w", err)
	}

	// TODO: Use database.SetClusterConfig() to store ClusterDNS and ClusterDomain
	kubeletArgs := []map[string]string{
		{"--cluster-dns": dnsIP},
		{"--cluster-domain": clusterDomain},
	}

	changed, err := snap.UpdateServiceArguments(s, "kubelet", kubeletArgs, []string{})
	if err != nil {
		return fmt.Errorf("failed to update 'kubelet' arguments: %w", err)
	}

	if changed {
		err = s.RestartService(ctx, "kubelet")
		if err != nil {
			return fmt.Errorf("failed to restart service 'kubelet': %w", err)
		}

	}
	return nil
}

func DisableDNSComponent(ctx context.Context, s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	err = manager.Disable("dns")
	if err != nil {
		return fmt.Errorf("failed to disable dns component: %w", err)
	}

	kubeletArgs := []map[string]string{
		{"--cluster-domain": "cluster.local"},
	}

	removeArgs := []string{"--cluster-dns"}

	changed, err := snap.UpdateServiceArguments(s, "kubelet", kubeletArgs, removeArgs)
	if err != nil {
		return fmt.Errorf("failed to update 'kubelet' arguments: %w", err)
	}

	if changed {
		err = s.RestartService(ctx, "kubelet")
		if err != nil {
			return fmt.Errorf("failed to restart service 'kubelet': %w", err)
		}

	}

	return nil
}
