package features

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/k8s"
)

func ApplyLoadBalancer(ctx context.Context, snap snap.Snap, cfg types.LoadBalancer) error {
	if !cfg.GetEnabled() {
		if err := disableLoadBalancer(ctx, snap); err != nil {
			return fmt.Errorf("failed to disable LoadBalancer: %w", err)
		}
		return nil
	}

	if err := enableLoadBalancer(ctx, snap, cfg); err != nil {
		return fmt.Errorf("failed to enable LoadBalancer: %w", err)
	}
	return nil
}

func disableLoadBalancer(ctx context.Context, snap snap.Snap) error {
	m := newHelm(snap)

	if _, err := m.Apply(ctx, featureLoadBalancer, stateDeleted, nil); err != nil {
		return fmt.Errorf("failed to uninstall LoadBalancer manifests: %w", err)
	}

	values := map[string]any{
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

	if _, err := m.Apply(ctx, featureNetwork, stateUpgradeOnly, values); err != nil {
		return fmt.Errorf("failed to refresh network to apply LoadBalancer configuration: %w", err)
	}
	return nil
}

func enableLoadBalancer(ctx context.Context, snap snap.Snap, cfg types.LoadBalancer) error {
	m := newHelm(snap)

	networkValues := map[string]any{
		"l2announcements": map[string]any{
			"enabled": cfg.GetL2Mode(),
		},
		"bgpControlPlane": map[string]any{
			"enabled": cfg.GetBGPMode(),
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

	changed, err := m.Apply(ctx, featureNetwork, stateUpgradeOnly, networkValues)
	if err != nil {
		return fmt.Errorf("failed to update Cilium configuration for LoadBalancer: %w", err)
	}

	if err := waitForRequiredLoadBalancerCRDs(ctx, snap, cfg.GetBGPMode()); err != nil {
		return fmt.Errorf("failed to wait for required Cilium CRDs to be available: %w", err)
	}

	cidrs := []map[string]any{}
	for _, cidr := range cfg.GetCIDRs() {
		cidrs = append(cidrs, map[string]any{"cidr": cidr})
	}
	for _, ipRange := range cfg.GetIPRanges() {
		cidrs = append(cidrs, map[string]any{"start": ipRange.Start, "stop": ipRange.Stop})
	}

	values := map[string]any{
		"l2": map[string]any{
			"enabled":    cfg.GetL2Mode(),
			"interfaces": cfg.GetL2Interfaces(),
		},
		"ipPool": map[string]any{
			"cidrs": cidrs,
		},
		"bgp": map[string]any{
			"enabled":  cfg.GetBGPMode(),
			"localASN": cfg.GetBGPLocalASN(),
			"neighbors": []map[string]any{
				{
					"peerAddress": cfg.GetBGPPeerAddress(),
					"peerASN":     cfg.GetBGPPeerASN(),
					"peerPort":    cfg.GetBGPPeerPort(),
				},
			},
		},
	}
	if _, err := m.Apply(ctx, featureLoadBalancer, statePresent, values); err != nil {
		return fmt.Errorf("failed to apply LoadBalancer configuration: %w", err)
	}

	if !changed {
		return nil
	}

	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart cilium to apply LoadBalancer configuration: %w", err)
	}
	return nil
}

func waitForRequiredLoadBalancerCRDs(ctx context.Context, snap snap.Snap, bgpMode bool) error {
	client, err := k8s.NewClient(snap.KubernetesRESTClientGetter(""))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return control.WaitUntilReady(ctx, func() (bool, error) {
		resources, err := client.ListResourcesForGroupVersion("cilium.io/v2alpha1")
		if err != nil {
			// This error is expected if the group version is not yet deployed.
			return false, nil
		}

		requiredCRDs := map[string]struct{}{
			"ciliuml2announcementpolicies": {},
			"ciliumloadbalancerippools":    {},
		}
		if bgpMode {
			requiredCRDs["ciliumbgppeeringpolicies"] = struct{}{}
		}
		requiredCount := len(requiredCRDs)
		for _, resource := range resources.APIResources {
			if _, ok := requiredCRDs[resource.Name]; ok {
				requiredCount = requiredCount - 1
			}
		}
		return requiredCount == 0, nil
	})
}
