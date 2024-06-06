package calico

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

// ApplyNetwork will deploy Calico when cfg.Enabled is true.
// ApplyNetwork will remove Calico when cfg.Enabled is false.
// ApplyNetwork returns an error if anything fails.
func ApplyNetwork(ctx context.Context, snap snap.Snap, cfg types.Network, _ types.Annotations) error {
	m := snap.HelmClient()

	if !cfg.GetEnabled() {
		if _, err := m.Apply(ctx, chartCalico, helm.StateDeleted, nil); err != nil {
			return fmt.Errorf("failed to uninstall network: %w", err)
		}
		return nil
	}

	podIpPools := []map[string]any{}
	ipv4PodCIDR, ipv6PodCIDR, err := utils.ParseCIDRs(cfg.GetPodCIDR())
	if err != nil {
		return fmt.Errorf("invalid pod cidr: %v", err)
	}
	if ipv4PodCIDR != "" {
		podIpPools = append(podIpPools, map[string]any{
			"name": "ipv4-ippool",
			"cidr": ipv4PodCIDR,
		})
	}
	if ipv6PodCIDR != "" {
		podIpPools = append(podIpPools, map[string]any{
			"name": "ipv6-ippool",
			"cidr": ipv6PodCIDR,
		})
	}

	serviceCIDRs := []string{}
	ipv4ServiceCIDR, ipv6ServiceCIDR, err := utils.ParseCIDRs(cfg.GetPodCIDR())
	if err != nil {
		return fmt.Errorf("invalid service cidr: %v", err)
	}
	if ipv4ServiceCIDR != "" {
		serviceCIDRs = append(serviceCIDRs, ipv4ServiceCIDR)
	}
	if ipv6ServiceCIDR != "" {
		serviceCIDRs = append(serviceCIDRs, ipv6ServiceCIDR)
	}

	values := map[string]any{
		"tigeraOperator": map[string]any{
			"registry": tigeraOperatorRegistry,
			"image":    tigeraOperatorImage,
			"version":  tigeraOperatorVersion,
		},
		"calicoctl": map[string]any{
			"image": calicoCtlImage,
			"tag":   calicoCtlTag,
		},
		"installation": map[string]any{
			"calicoNetwork": map[string]any{
				"ipPools": podIpPools,
			},
		},
		"serviceCIDRs": serviceCIDRs,
	}

	if _, err := m.Apply(ctx, chartCalico, helm.StatePresent, values); err != nil {
		return fmt.Errorf("failed to enable network: %w", err)
	}

	return nil
}
