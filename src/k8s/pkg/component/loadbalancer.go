package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/k8s/pkg/utils/vals"
)

func UpdateLoadBalancerComponent(ctx context.Context, s snap.Snap, isRefresh bool, config types.LoadBalancer) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	networkValues := map[string]any{
		"l2announcements": map[string]any{
			"enabled": config.GetL2Mode(),
		},
		"bgpControlPlane": map[string]any{
			"enabled": config.GetBGPMode(),
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
		return fmt.Errorf("failed to enable load-balancer component: %w", err)
	}

	// Wait for cilium CRDs to be available.
	k8sClient, err := k8s.NewClient(s.KubernetesRESTClientGetter(""))
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		resources, err := k8sClient.ListResourcesForGroupVersion("cilium.io/v2alpha1")
		if err != nil {
			// This error is expected if the group version is not yet deployed.
			return false, nil
		}

		requiredCRDs := map[string]struct{}{
			"ciliuml2announcementpolicies": {},
			"ciliumloadbalancerippools":    {},
		}
		if config.GetBGPMode() {
			requiredCRDs["ciliumbgppeeringpolicies"] = struct{}{}
		}
		requiredCount := len(requiredCRDs)
		for _, resource := range resources.APIResources {
			if _, ok := requiredCRDs[resource.Name]; ok {
				requiredCount = requiredCount - 1
			}
		}
		return requiredCount == 0, nil
	}); err != nil {
		return fmt.Errorf("failed to wait for cilium CRDs to be available: %w", err)
	}

	formattedCidrs := []map[string]any{}

	for _, cidr := range config.GetCIDRs() {
		// Handle IP range
		if strings.Contains(cidr, "-") {
			ipRange := strings.Split(cidr, "-")
			formattedCidrs = append(formattedCidrs, map[string]any{"start": ipRange[0], "stop": ipRange[1]})
		} else {
			// Handle CIDRs
			formattedCidrs = append(formattedCidrs, map[string]any{"cidr": cidr})
		}
	}

	values := map[string]any{
		"l2": map[string]any{
			"enabled":    config.GetL2Mode(),
			"interfaces": config.GetL2Interfaces(),
		},
		"ipPool": map[string]any{
			"cidrs": formattedCidrs,
		},
		"bgp": map[string]any{
			"enabled":  config.GetBGPMode(),
			"localASN": config.GetBGPLocalASN(),
			"neighbors": []map[string]any{
				{
					"peerAddress": config.GetBGPPeerAddress(),
					"peerASN":     config.GetBGPPeerASN(),
					"peerPort":    config.GetBGPPeerPort(),
				},
			},
		},
	}

	if isRefresh {
		if err := manager.Refresh("load-balancer", values); err != nil {
			return fmt.Errorf("failed to refresh load-balancer component: %w", err)
		}
	} else {
		if err := manager.Enable("load-balancer", values); err != nil {
			return fmt.Errorf("failed to enable load-balancer component: %w", err)
		}
	}

	client, err := k8s.NewClient(s.KubernetesRESTClientGetter(""))
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	attempts := 3
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

func DisableLoadBalancerComponent(s snap.Snap) error {
	manager, err := NewHelmClient(s, nil)
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	if err := manager.Disable("load-balancer"); err != nil {
		return fmt.Errorf("failed to disable load-balancer component: %w", err)
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

func ReconcileLoadBalancerComponent(ctx context.Context, s snap.Snap, alreadyEnabled *bool, requestEnabled *bool, clusterConfig types.ClusterConfig) error {
	if vals.OptionalBool(requestEnabled, true) && vals.OptionalBool(alreadyEnabled, false) {
		// If already enabled, and request does not contain `enabled` key
		// or if already enabled and request contains `enabled=true`
		err := UpdateLoadBalancerComponent(ctx, s, true, clusterConfig.LoadBalancer)
		if err != nil {
			return fmt.Errorf("failed to refresh load-balancer: %w", err)
		}
		return nil
	} else if vals.OptionalBool(requestEnabled, false) {
		err := UpdateLoadBalancerComponent(ctx, s, false, clusterConfig.LoadBalancer)
		if err != nil {
			return fmt.Errorf("failed to enable load-balancer: %w", err)
		}
		return nil
	} else if !vals.OptionalBool(requestEnabled, false) {
		err := DisableLoadBalancerComponent(s)
		if err != nil {
			return fmt.Errorf("failed to disable load-balancer: %w", err)
		}
		return nil
	}
	return nil
}
