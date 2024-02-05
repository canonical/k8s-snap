package mocks

import "helm.sh/helm/v3/pkg/action"

// "Mock" Initializer
type HelmClientInitializer struct {
	ActionConfig *action.Configuration
}

func (r *HelmClientInitializer) InitializeHelmClientConfig() (*action.Configuration, error) {
	return r.ActionConfig, nil
}
