package network

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	CiliumChartName = "cilium"

	CiliumAgentImageName    = "cilium-agent"
	CiliumOperatorImageName = "cilium-operator"
)

var manifest = types.FeatureManifest{
	Name:    "network",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		CiliumChartName: {
			Name:             "cilium",
			Version:          "1.17.1",
			InstallName:      "ck-network",
			InstallNamespace: "kube-system",
		},
	},

	Images: map[string]types.Image{
		"cilium-agent": {
			Registry:   "ghcr.io/canonical",
			Repository: "cilium",
			Tag:        "1.17.1-ck0",
		},
		"cilium-operator": {
			Registry:   "ghcr.io/canonical",
			Repository: "cilium-operator",
			Tag:        "1.17.1-ck0",
		},
	},
}

var FeatureNetwork types.Feature = manifest
