package component

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
)

func EnableLoadBalancerComponent(s snap.Snap, cidrs []string, l2Enabled bool, l2Interfaces []string, bgpEnabled bool, bgpLocalASN int, bgpPeerAddress string, bgpPeerASN int, bgpPeerPort int) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	networkValues := map[string]any{
		"l2announcements": map[string]any{
			"enabled": l2Enabled,
		},
		"bgpControlPlane": map[string]any{
			"enabled": bgpEnabled,
		},
		"externalIPs": map[string]any{
			"enabled": true,
		},
		// https://docs.cilium.io/en/v1.14/network/l2-announcements/#sizing-client-rate-limit
		// Assuming for 50 LB services
		"k8sClientRateLimit": map[string]any{
			"qps":   10,
			"burst": 20,
		},
	}

	if err := manager.Refresh("network", networkValues); err != nil {
		return fmt.Errorf("failed to enable ingress component: %w", err)
	}

	formattedCidrs := []map[string]any{}

	for _, cidr := range cidrs {
		formattedCidrs = append(formattedCidrs, map[string]any{"cidr": cidr})
	}

	values := map[string]any{
		"l2": map[string]any{
			"enabled":    l2Enabled,
			"interfaces": l2Interfaces,
		},
		"ipPool": map[string]any{
			"cidrs": formattedCidrs,
		},
		"bgp": map[string]any{
			"enabled":  bgpEnabled,
			"localASN": bgpLocalASN,
			"neighbors": []map[string]any{
				map[string]any{
					"peerAddress": bgpPeerAddress,
					"peerASN":     bgpPeerASN,
					"peerPort":    bgpPeerPort,
				},
			},
		},
	}

	if err := manager.Enable("loadbalancer", values); err != nil {
		return fmt.Errorf("failed to enable loadbalancer component: %w", err)
	}

	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := k8s.RestartDeployment(ctx, client, "cilium-operator", "kube-system"); err != nil {
		return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
	}

	if err := k8s.RestartDaemonset(ctx, client, "cilium", "kube-system"); err != nil {
		return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
	}

	return nil
}

func DisableLoadBalancerComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	if err := manager.Disable("loadbalancer"); err != nil {
		return fmt.Errorf("failed to disable loadbalancer component: %w", err)
	}

	networkValues := map[string]any{
		"l2announcements": map[string]any{
			"enabled": false,
		},
		"bgpControlPlane": map[string]any{
			"enabled": false,
		},
		"externalIPs": map[string]any{
			"enabled": false,
		},
		// https://docs.cilium.io/en/v1.14/network/l2-announcements/#sizing-client-rate-limit
		// Setting back to defaults
		"k8sClientRateLimit": map[string]any{
			"qps":   5,
			"burst": 10,
		},
	}

	if err := manager.Refresh("network", networkValues); err != nil {
		return fmt.Errorf("failed to disable ingress component: %w", err)
	}

	return nil
}
