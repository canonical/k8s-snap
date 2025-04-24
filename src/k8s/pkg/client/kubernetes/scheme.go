package kubernetes

import (
	"fmt"

	crdsv1 "github.com/canonical/k8s/pkg/k8sd/crds/api/v1alpha"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewScheme creates a new runtime.Scheme and adds the types defined in this package to it.
func NewScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := crdsv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add types to scheme: %w", err)
	}
	return scheme, nil
}
