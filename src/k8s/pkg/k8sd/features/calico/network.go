package calico

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

const (
	EnabledMsg          = "enabled"
	DisabledMsg         = "disabled"
	deployFailedMsgTmpl = "Failed to deploy Calico, the error was: %v"
	deleteFailedMsgTmpl = "Failed to delete Calico, the error was: %v"
)

// ApplyNetwork will deploy Calico when network.Enabled is true.
// ApplyNetwork will remove Calico when network.Enabled is false.
// ApplyNetwork will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyNetwork returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyNetwork(ctx context.Context, snap snap.Snap, _ string, apiserver types.APIServer, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	if !network.GetEnabled() {
		if _, err := m.Apply(ctx, ChartCalico, helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall network: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CalicoTag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, err
		}

		return types.FeatureStatus{
			Enabled: false,
			Version: CalicoTag,
			Message: DisabledMsg,
		}, nil
	}

	config, err := internalConfig(annotations)
	if err != nil {
		err = fmt.Errorf("failed to parse annotations: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CalicoTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	podIpPools := []map[string]any{}
	ipv4PodCIDR, ipv6PodCIDR, err := utils.SplitCIDRStrings(network.GetPodCIDR())
	if err != nil {
		err = fmt.Errorf("invalid pod cidr: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CalicoTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}
	if ipv4PodCIDR != "" {
		podIpPools = append(podIpPools, map[string]any{
			"name":          "ipv4-ippool",
			"cidr":          ipv4PodCIDR,
			"encapsulation": config.encapsulationV4,
		})
	}
	if ipv6PodCIDR != "" {
		podIpPools = append(podIpPools, map[string]any{
			"name":          "ipv6-ippool",
			"cidr":          ipv6PodCIDR,
			"encapsulation": config.encapsulationV6,
		})
	}

	serviceCIDRs := []string{}
	ipv4ServiceCIDR, ipv6ServiceCIDR, err := utils.SplitCIDRStrings(network.GetServiceCIDR())
	if err != nil {
		err = fmt.Errorf("invalid service cidr: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CalicoTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}
	if ipv4ServiceCIDR != "" {
		serviceCIDRs = append(serviceCIDRs, ipv4ServiceCIDR)
	}
	if ipv6ServiceCIDR != "" {
		serviceCIDRs = append(serviceCIDRs, ipv6ServiceCIDR)
	}

	calicoNetworkValues := map[string]any{
		"ipPools": podIpPools,
	}

	if config.autodetectionV4 != nil {
		calicoNetworkValues["nodeAddressAutodetectionV4"] = config.autodetectionV4
	}

	if config.autodetectionV6 != nil {
		calicoNetworkValues["nodeAddressAutodetectionV6"] = config.autodetectionV6
	}

	values := map[string]any{
		"tigeraOperator": map[string]any{
			"registry": imageRepo,
			"image":    tigeraOperatorImage,
			"version":  tigeraOperatorVersion,
		},
		"calicoctl": map[string]any{
			"image": calicoCtlImage,
			"tag":   calicoCtlTag,
		},
		"installation": map[string]any{
			"calicoNetwork": calicoNetworkValues,
			"registry":      imageRepo,
		},
		"apiServer": map[string]any{
			"enabled": config.apiServerEnabled,
		},
		"serviceCIDRs": serviceCIDRs,
	}

	if _, err := m.Apply(ctx, ChartCalico, helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to enable network: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CalicoTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: CalicoTag,
		Message: EnabledMsg,
	}, nil
}
