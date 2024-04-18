package features

import "context"

type Manager interface {
	// Apply applies a feature f on the cluster.
	Apply(ctx context.Context, f feature, desired state, values map[string]any) (bool, error)
}

type feature struct {
	// name is the user-recognisable feature name.
	name string

	// namespace is the namespace to install the feature.
	namespace string

	// manifestPath is the path of the feature manifests.
	manifestPath string
}
