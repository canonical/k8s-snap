package mock

import (
	"context"

	"github.com/canonical/k8s/pkg/client/helm"
)

type MockApplyArguments struct {
	Context context.Context
	Chart   helm.InstallableChart
	State   helm.State
	Values  map[string]any
}

// Mock is a mock implementation of helm.Client
type Mock struct {
	ApplyCalledWith []MockApplyArguments
	ApplyChanged    bool
	ApplyErr        error
}

// Apply implements helm.Client
func (m *Mock) Apply(ctx context.Context, c helm.InstallableChart, desired helm.State, values map[string]any) (bool, error) {
	m.ApplyCalledWith = append(m.ApplyCalledWith, MockApplyArguments{Context: ctx, Chart: c, State: desired, Values: values})
	return m.ApplyChanged, m.ApplyErr
}

var _ helm.Client = &Mock{}
