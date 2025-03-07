package ingress

import (
	"fmt"

	"dario.cat/mergo"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type CiliumValues cilium_network.Values

func (v CiliumValues) applyDefaultValues() error {
	values := map[string]any{
		"ingressController": map[string]any{
			"loadbalancerMode":       "shared",
			"defaultSecretNamespace": "kube-system",
		},
	}

	if err := mergo.Merge(&v, CiliumValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v CiliumValues) applyClusterConfiguration(ingress types.Ingress) error {
	values := map[string]any{
		"ingressController": map[string]any{
			IngressOptionEnabled:             ingress.GetEnabled(),
			IngressOptionDefaultSecretName:   ingress.GetDefaultTLSSecret(),
			IngressOptionEnableProxyProtocol: ingress.GetEnableProxyProtocol(),
		},
	}

	if err := mergo.Merge(&v, CiliumValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge cluster configuration: %w", err)
	}

	return nil
}

func (v CiliumValues) applyDisableConfiguration() error {
	values := map[string]any{
		"ingressController": map[string]any{
			IngressOptionEnabled:                false,
			IngressOptionLoadBalancerMode:       "",
			IngressOptionDefaultSecretNamespace: "",
			IngressOptionDefaultSecretName:      "",
			IngressOptionEnableProxyProtocol:    false,
		},
	}

	if err := mergo.Merge(&v, CiliumValues(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge disable configuration: %w", err)
	}

	return nil
}
