package component

import "helm.sh/helm/v3/pkg/action"

// ComponentManager defines an interface for managing k8s components.
type ComponentManager interface {
	// Enable enables a k8s component, optionally specifying custom configuration options.
	Enable(name string, values map[string]any) error
	// List returns a list of enabled components.
	List() ([]Component, error)
	// Disable disables a component from the cluster.
	Disable(name string) error
	// Refresh updates a k8s component.
	Refresh(name string) error
}

// HelmConfigInitializer defines an interface for initializing a Helm Configuration, allowing a Mock implementation
type HelmConfigInitializer interface {
	// Initializes a fresh Helm Configuration
	New() (*action.Configuration, error)
}
