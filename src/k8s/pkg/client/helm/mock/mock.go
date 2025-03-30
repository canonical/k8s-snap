package mock

import (
	"context"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type MockApplyArguments struct {
	Context context.Context
	Feature types.FeatureName
	Version string
	Chart   helm.InstallableChart
	State   helm.State
	Values  map[string]any
}

// Mock is a mock implementation of helm.Client.
type Mock struct {
	ApplyCalledWith []MockApplyArguments
	ApplyChanged    bool
	ApplyErr        error
}

// Apply implements helm.Client.
func (m *Mock) Apply(ctx context.Context, feature types.FeatureName, version string, c helm.InstallableChart, desired helm.State, values map[string]any) (bool, error) {
	m.ApplyCalledWith = append(m.ApplyCalledWith, MockApplyArguments{Context: ctx, Feature: feature, Version: version, Chart: c, State: desired, Values: values})
	return m.ApplyChanged, m.ApplyErr
}

// Apply implements helm.Client.
func (m *Mock) ApplyDependent(ctx context.Context, parent helm.FeatureMeta, sub helm.PseudoFeatureMeta, desired helm.State, values map[string]any) (bool, error) {
	m.ApplyCalledWith = append(m.ApplyCalledWith, MockApplyArguments{Context: ctx, Feature: sub.FeatureName, Version: sub.Version, Chart: parent.Chart, State: desired, Values: values})
	return m.ApplyChanged, m.ApplyErr
}

var _ helm.Client = &Mock{}
