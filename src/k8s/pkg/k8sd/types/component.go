package types

// Component defines a Kubernetes component that can be deployed on the cluster.
type Component struct {
	// DependsOn is a component that this component depends on.
	DependsOn string
	// ManifestPath is a path containing the manifests to deploy this component.
	ManifestPath string
	// ReleaseName is the name to use when applying this component on the cluster.
	ReleaseName string
	// Namespace is the namespace where this component is installed.
	Namespace string
}
