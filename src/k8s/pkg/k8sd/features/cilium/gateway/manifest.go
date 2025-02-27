package gateway

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	GatewayChartName      = "gateway-api"
	GatewayClassChartName = "ck-gateway-cilium"
)

var manifest = types.FeatureManifest{
	Name:    "gateway",
	Version: "1.0.0",

	Charts: map[string]helm.InstallableChart{
		GatewayChartName: {
			Name:             "gateway-api",
			Version:          "1.2.0",
			InstallName:      "ck-gateway",
			InstallNamespace: "kube-system",
		},
		GatewayClassChartName: {
			Name:             "ck-gateway-cilium",
			Version:          "0.1.0",
			InstallName:      "ck-gateway-class",
			InstallNamespace: "default",
		},
	},
}

var FeatureGateway types.Feature = manifest
