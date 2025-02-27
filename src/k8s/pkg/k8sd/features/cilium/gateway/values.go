package gateway

import (
	"fmt"

	"dario.cat/mergo"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
)

type CiliumValues cilium_network.Values

func (v CiliumValues) ApplyClusterConfiguration() error {
	values := map[string]any{
		"gatewayAPI": map[string]any{
			// TODO(berkayoz): This can be fetched from gateway.GetEnabled()
			"enabled": true,
		},
	}

	if err := mergo.Merge(&v, CiliumValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge cluster configuration: %w", err)
	}

	return nil
}

func (v CiliumValues) ApplyDisableConfiguration() error {
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
