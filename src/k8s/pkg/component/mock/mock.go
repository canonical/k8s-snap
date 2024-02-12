package mock

import "helm.sh/helm/v3/pkg/action"

// "Mock" Initializer
type MockHelmConfigProvider struct {
	ActionConfig *action.Configuration
}

func (r *MockHelmConfigProvider) New(namespace string) (*action.Configuration, error) {
	return r.ActionConfig, nil
}
