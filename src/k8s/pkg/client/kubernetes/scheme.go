package kubernetes

import (
	"fmt"

	upgradesv1alpha "github.com/canonical/k8s/pkg/k8sd/crds/upgrades/v1alpha"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// NewScheme creates a new runtime.Scheme and adds the types defined in this package to it.
// It also adds the client-go scheme to the scheme.
func NewScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add client-go types to scheme: %w", err)
	}
	if err := upgradesv1alpha.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add types to scheme: %w", err)
	}
	return scheme, nil
}
