package manifests

import "github.com/canonical/k8s/pkg/k8sd/types"

// registeredManifests is a list of feature manifests bundled in k8s-snap.
var registeredManifests []*types.FeatureManifest

// Manifests returns the list of registered manifests.
func Manifests() []*types.FeatureManifest {
	if registeredManifests == nil {
		return nil
	}
	manifestList := make([]*types.FeatureManifest, len(registeredManifests))
	copy(manifestList, registeredManifests)
	return manifestList
}

// Register manifests that are used by k8s-snap.
// Register is used by the `init()` method in individual packages.
func Register(charts *types.FeatureManifest) {
	registeredManifests = append(registeredManifests, charts)
}
