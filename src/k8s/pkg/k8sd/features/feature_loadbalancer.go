package features

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
)

// ApplyLoadBalancer is used to configure the load-balancer feature on Canonical Kubernetes.
// ApplyLoadBalancer assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyLoadBalancer will configure Cilium to enable L2 or BGP mode, and deploy necessary CRs for announcing the LoadBalancer external IPs when loadbalancer.Enabled is true.
// ApplyLoadBalancer will disable L2 and BGP on Cilium, and remove any previously created CRs when loadbalancer.Enabled is false.
// ApplyLoadBalancer will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyLoadBalancer returns an error if anything fails.
func ApplyLoadBalancer(ctx context.Context, snap snap.Snap, loadbalancer types.LoadBalancer, network types.Network) error {
	if !loadbalancer.GetEnabled() {
		if err := disableLoadBalancer(ctx, snap, network); err != nil {
			return fmt.Errorf("failed to disable LoadBalancer: %w", err)
		}
		return nil
	}

	if err := enableLoadBalancer(ctx, snap, loadbalancer, network); err != nil {
		return fmt.Errorf("failed to enable LoadBalancer: %w", err)
	}
	return nil
}

func disableLoadBalancer(ctx context.Context, snap snap.Snap, network types.Network) error {
	m := snap.HelmClient()

	if _, err := m.Apply(ctx, chartCiliumLoadBalancer, helm.StateDeleted, nil); err != nil {
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

	if _, err := m.Apply(ctx, chartCilium, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), values); err != nil {
		return fmt.Errorf("failed to refresh network to apply LoadBalancer configuration: %w", err)
	}
	return nil
}

func enableLoadBalancer(ctx context.Context, snap snap.Snap, loadbalancer types.LoadBalancer, network types.Network) error {
	m := snap.HelmClient()

	networkValues := map[string]any{
		"l2announcements": map[string]any{
			"enabled": loadbalancer.GetL2Mode(),
		},
		"bgpControlPlane": map[string]any{
			"enabled": loadbalancer.GetBGPMode(),
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

	changed, err := m.Apply(ctx, chartCilium, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), networkValues)
	if err != nil {
		return fmt.Errorf("failed to update Cilium configuration for LoadBalancer: %w", err)
	}

	if err := waitForRequiredLoadBalancerCRDs(ctx, snap, loadbalancer.GetBGPMode()); err != nil {
		return fmt.Errorf("failed to wait for required Cilium CRDs to be available: %w", err)
	}

	cidrs := []map[string]any{}
	for _, cidr := range loadbalancer.GetCIDRs() {
		cidrs = append(cidrs, map[string]any{"cidr": cidr})
	}
	for _, ipRange := range loadbalancer.GetIPRanges() {
		cidrs = append(cidrs, map[string]any{"start": ipRange.Start, "stop": ipRange.Stop})
	}

	values := map[string]any{
		"l2": map[string]any{
			"enabled":    loadbalancer.GetL2Mode(),
			"interfaces": loadbalancer.GetL2Interfaces(),
		},
		"ipPool": map[string]any{
			"cidrs": cidrs,
		},
		"bgp": map[string]any{
			"enabled":  loadbalancer.GetBGPMode(),
			"localASN": loadbalancer.GetBGPLocalASN(),
			"neighbors": []map[string]any{
				{
					"peerAddress": loadbalancer.GetBGPPeerAddress(),
					"peerASN":     loadbalancer.GetBGPPeerASN(),
					"peerPort":    loadbalancer.GetBGPPeerPort(),
				},
			},
		},
	}
	if _, err := m.Apply(ctx, chartCiliumLoadBalancer, helm.StatePresent, values); err != nil {
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
	client, err := snap.KubernetesClient("")
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
