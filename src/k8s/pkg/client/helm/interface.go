package helm

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
)

type FeatureMeta struct {
	FeatureName types.FeatureName
	Version     string
	Chart       InstallableChart
}

type PseudoFeatureMeta struct {
	FeatureName types.FeatureName
	Version     string
}

// Client handles the lifecycle of charts (manifests + config) on the cluster.
type Client interface {
	// Apply ensures the state of a InstallableChart on the cluster.
	// When state is StatePresent, Apply will install or upgrade the chart using the specified values as configuration. Apply returns true if the chart was not installed, or any values were changed.
	// When state is StateUpgradeOnly, Apply will upgrade the chart using the specified values as configuration. Apply returns true if the chart was not installed, or any values were changed. An error is returned if the chart is not already installed.
	// When state is StateDeleted, Apply will ensure that the chart is removed. If the chart is not installed, this is a no-op. Apply returns true if the chart was previously installed.
	// Apply returns an error in case of failure.
	Apply(ctx context.Context, feature types.FeatureName, version string, f InstallableChart, desired State, values map[string]any) (bool, error)
	ApplyDependent(ctx context.Context, parent FeatureMeta, sub PseudoFeatureMeta, desired State, values map[string]any) (bool, error)
}
