package cilium

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type ingressValues map[string]any

func (v *ingressValues) applyDefaults() error {
	values := ingressValues{
		"ingressController": map[string]any{
			IngressOptionEnabled:                true,
			IngressOptionLoadBalancerMode:       IngressOptionLoadBalancerModeShared,
			IngressOptionDefaultSecretNamespace: IngressOptionDefaultSecretNamespaceKubeSystem,
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *ingressValues) applyClusterConfig(ingress types.Ingress) error {
	values := ingressValues{
		"ingressController": map[string]any{
			IngressOptionDefaultSecretName:   ingress.GetDefaultTLSSecret(),
			IngressOptionEnableProxyProtocol: ingress.GetEnableProxyProtocol(),
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *ingressValues) applyDisable() error {
	values := ingressValues{
		"ingressController": map[string]any{
			IngressOptionEnabled:                false,
			IngressOptionLoadBalancerMode:       "",
			IngressOptionDefaultSecretNamespace: "",
			IngressOptionDefaultSecretName:      "",
			IngressOptionEnableProxyProtocol:    false,
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}
