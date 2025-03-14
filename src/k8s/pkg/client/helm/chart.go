package helm

// InstallableChart describes a chart that can be deployed on a running cluster.
type InstallableChart struct {
	// Name is the install name of the chart.
	Name string

	// Namespace is the namespace to install the chart.
	Namespace string

	// ManifestPath is the path to the chart's manifest, typically relative to "$SNAP/k8s/manifests".
	// TODO(neoaggelos): this should be a *chart.Chart, and we should use the "embed" package to load it when building k8sd.
	ManifestPath string
}
