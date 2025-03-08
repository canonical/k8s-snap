package network

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	CalicoChartName = "tigera-operator"

	CalicoImageName         = "calico"
	TigeraOperatorImageName = "tigera-operator"
	CalicoCtlImageName      = "calicoctl"
)

var manifest = types.FeatureManifest{
	Name:    "network",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		CalicoChartName: {
			Name:             "tigera-operator",
			Version:          "v3.28.0",
			InstallName:      "ck-network",
			InstallNamespace: "tigera-operator",
		},
	},

	Images: map[string]types.Image{
		"calico": {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "calico",
			Tag:        "v3.28.0",
		},
		"tigera-operator": {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "tigera/operator",
			Tag:        "v1.34.0",
		},
		"calicoctl": {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "calico/ctl",
			Tag:        "v3.28.0",
		},
	},
}

var FeatureNetwork types.Feature = manifest
