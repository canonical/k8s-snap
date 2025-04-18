package kubernetes

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// NewScheme creates a new runtime.Scheme and adds the types defined in this package to it.
func NewScheme() (*runtime.Scheme, error) {
	schemeBuilder := runtime.NewSchemeBuilder(
		addUpgradeTypes,
		// NOTE(Hue): Add other types here as needed.
	)
	addToScheme := schemeBuilder.AddToScheme

	scheme := runtime.NewScheme()
	if err := addToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add types to scheme: %w", err)
	}
	return scheme, nil
}
