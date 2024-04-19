package features

import "context"

// Manager handles the lifecycle of features (manifests + config) on the cluster.
type Manager interface {
	// Apply ensures the state of a Feature on the cluster.
	// When state is statePresent, Apply will install or upgrade the feature using the specified values as configuration. Apply returns true if the feature was not installed, or any values were changed.
	// When state is stateUpgradeOnly, Apply will upgrade the feature using the specified values as configuration. Apply returns true if the feature was not installed, or any values were changed. An error is returned if the feature is not already installed.
	// When state is stateDeleted, Apply will ensure that the feature is removed. If the feature is not installed, this is a no-op. Apply returns true if the feature was previously installed.
	// Apply returns an error in case of failure.
	Apply(ctx context.Context, f Feature, desired state, values map[string]any) (bool, error)
}

// feature describes a feature that can be deployed on a running cluster.
type Feature struct {
	// name is the install name of the feature.
	name string

	// namespace is the namespace to install the feature.
	namespace string

	// manifestPath is the path to the feature's manifest, relative to the Snap.ManifestsDir(), typically "$SNAP/k8s/manifests".
	// TODO(neoaggelos): this should be a *chart.Chart, and we should use the "embed" package to load it during build.
	manifestPath string
}
