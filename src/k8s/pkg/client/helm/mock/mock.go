package mock

import (
	"context"
	"slices"
	"sync"

	"github.com/canonical/k8s/pkg/client/helm"
)

type MockApplyArguments struct {
	Context context.Context
	Chart   helm.InstallableChart
	State   helm.State
	Values  map[string]any
}

// Mock is a mock implementation of helm.Client.
type Mock struct {
	mu              sync.Mutex
	ApplyCalledWith []MockApplyArguments
	ApplyChanged    bool
	ApplyErr        error
}

// Apply implements helm.Client.
func (m *Mock) Apply(ctx context.Context, c helm.InstallableChart, desired helm.State, values map[string]any) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ApplyCalledWith = append(m.ApplyCalledWith, MockApplyArguments{Context: ctx, Chart: c, State: desired, Values: values})
	return m.ApplyChanged, m.ApplyErr
}

// Calls returns a shallow copy of the recorded Apply invocations slice.
func (m *Mock) Calls() []MockApplyArguments {
	m.mu.Lock()
	defer m.mu.Unlock()

	return slices.Clone(m.ApplyCalledWith)
}

var _ helm.Client = &Mock{}
