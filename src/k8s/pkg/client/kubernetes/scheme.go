package kubernetes

import (
	"fmt"

	upgradesv1alpha "github.com/canonical/k8s/pkg/k8sd/crds/upgrades/v1alpha"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewScheme creates a new runtime.Scheme and adds the types defined in this package to it.
func NewScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := upgradesv1alpha.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add types to scheme: %w", err)
	}
	return scheme, nil
}
