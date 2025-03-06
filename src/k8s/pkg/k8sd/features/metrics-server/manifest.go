package metrics_server

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	MetricsServerChartName = "metrics-server"

	MetricsServerImageName = "metrics-server"
)

var Manifest = types.FeatureManifest{
	Name:    "metrics-server",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		MetricsServerChartName: {
			Name:             "metrics-server",
			Version:          "3.12.2",
			InstallName:      "metrics-server",
			InstallNamespace: "kube-system",
		},
	},

	Images: map[string]types.Image{
		MetricsServerImageName: {
			Registry:   "ghcr.io/canonical",
			Repository: "metrics-server",
			Tag:        "0.7.2-ck0",
		},
	},
}
