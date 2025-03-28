package cilium

import (
	"fmt"

	"dario.cat/mergo"
)

type gatewayValues map[string]any

func (v *gatewayValues) applyDefaults() error {
	values := gatewayValues{"gatewayAPI": map[string]any{"enabled": true}}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *gatewayValues) applyDisable() error {
	values := gatewayValues{"gatewayAPI": map[string]any{"enabled": false}}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}
