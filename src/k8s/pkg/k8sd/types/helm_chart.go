package types

type HelmChart struct {
	// Name is the name of the chart.
	Name string `json:"name"`
	// Version is the version of the chart.
	Version string `json:"version"`
	// Contents is the contents of the chart.
	Contents []byte `json:"contents"`
}
