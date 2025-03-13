package network

import (
	"context"
	"fmt"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
)

type Values map[string]any

func (v Values) applyDefaultValues() error {
	values := map[string]any{
		"apiServer": map[string]any{
			"enabled": false,
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v Values) ApplyImageOverrides() error {
	tigeraOperatorImage := FeatureNetwork.GetImage(TigeraOperatorImageName)
	calicoCtlImage := FeatureNetwork.GetImage(CalicoCtlImageName)
	calicoImage := FeatureNetwork.GetImage(CalicoImageName)

	values := map[string]any{
		"tigeraOperator": map[string]any{
			"registry": tigeraOperatorImage.Registry,
			"image":    tigeraOperatorImage.Repository,
			"version":  tigeraOperatorImage.Tag,
		},
		"calicoctl": map[string]any{
			"image": calicoCtlImage.GetURI(),
			"tag":   calicoCtlImage.Tag,
		},
		"installation": map[string]any{
			"registry": calicoImage.Registry,
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v Values) applyClusterConfiguration(ctx context.Context, s state.State, apiserver types.APIServer, network types.Network) error {
	podIpPools := []map[string]any{}
	ipv4PodCIDR, ipv6PodCIDR, err := utils.SplitCIDRStrings(network.GetPodCIDR())
	if err != nil {
		return fmt.Errorf("invalid pod cidr: %w", err)
	}
	if ipv4PodCIDR != "" {
		podIpPools = append(podIpPools, map[string]any{
			"name":          "ipv4-ippool",
			"cidr":          ipv4PodCIDR,
			"encapsulation": "VXLAN",
		})
	}
	if ipv6PodCIDR != "" {
		podIpPools = append(podIpPools, map[string]any{
			"name":          "ipv6-ippool",
			"cidr":          ipv6PodCIDR,
			"encapsulation": "VXLAN",
		})
	}

	serviceCIDRs := []string{}
	ipv4ServiceCIDR, ipv6ServiceCIDR, err := utils.SplitCIDRStrings(network.GetServiceCIDR())
	if err != nil {
		return fmt.Errorf("invalid service cidr: %w", err)
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

	values := map[string]any{
		"installation": map[string]any{
			"calicoNetwork": calicoNetworkValues,
		},
		"serviceCIDRs": serviceCIDRs,
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v Values) ApplyAnnotations(annotations types.Annotations) error {
	config, err := internalConfig(annotations)
	if err != nil {
		return fmt.Errorf("failed to parse annotations: %w", err)
	}

	// Merging slice of maps with mergo.Merge depends on the order of the elements in the slice.
	// Instead of using mergo.Merge, we will manually merge the values.
	installation, ok := v["installation"].(map[string]any)
	if !ok {
		return fmt.Errorf("installation values not found")
	}

	calicoNetwork, ok := installation["calicoNetwork"].(map[string]any)
	if !ok {
		return fmt.Errorf("calicoNetwork values not found")
	}

	ipPools, ok := calicoNetwork["ipPools"].([]map[string]any)
	if !ok {
		return fmt.Errorf("ipPools values not found")
	}

	for _, pool := range ipPools {
		if config.encapsulationV4 != "" && pool["name"] == "ipv4-ippool" {
			pool["encapsulation"] = config.encapsulationV4
		}
		if config.encapsulationV6 != "" && pool["name"] == "ipv6-ippool" {
			pool["encapsulation"] = config.encapsulationV6
		}
	}

	calicoNetworkValues := map[string]any{}

	if config.autodetectionV4 != nil {
		calicoNetworkValues["nodeAddressAutodetectionV4"] = config.autodetectionV4
	}

	if config.autodetectionV6 != nil {
		calicoNetworkValues["nodeAddressAutodetectionV6"] = config.autodetectionV6
	}

	values := map[string]any{
		"installation": map[string]any{
			"calicoNetwork": calicoNetworkValues,
		},
		"apiServer": map[string]any{
			"enabled": config.apiServerEnabled,
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}
