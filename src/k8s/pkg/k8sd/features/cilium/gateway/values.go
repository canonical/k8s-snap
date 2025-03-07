package gateway

import (
	"fmt"

	"dario.cat/mergo"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type CiliumValues cilium_network.Values

func (v CiliumValues) applyClusterConfiguration(gateway types.Gateway) error {
	values := map[string]any{
		"gatewayAPI": map[string]any{
			"enabled": gateway.GetEnabled(),
		},
	}

	if err := mergo.Merge(&v, CiliumValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge cluster configuration: %w", err)
	}

	return nil
}

func (v CiliumValues) applyDisableConfiguration() error {
	values := map[string]any{
		"gatewayAPI": map[string]any{
			"enabled": false,
		},
	}

	if err := mergo.Merge(&v, CiliumValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge disable configuration: %w", err)
	}

	return nil
}
