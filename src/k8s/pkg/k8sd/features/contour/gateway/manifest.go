package gateway

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	ChartGatewayName = "ck-gateway-contour"

	ContourGatewayProvisionerEnvoyImageName   = "envoy-gateway"
	ContourGatewayProvisionerContourImageName = "contour-gateway"
)

var manifest = types.FeatureManifest{
	Name:    "gateway",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		ChartGatewayName: {
			Name:             "ck-gateway-contour",
			Version:          "1.28.2",
			InstallName:      "ck-gateway",
			InstallNamespace: "projectcontour",
		},
	},

	Images: map[string]types.Image{
		ContourGatewayProvisionerContourImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "projectcontour/contour",
			Tag:        "v1.28.2",
		},
		ContourGatewayProvisionerEnvoyImageName: {
			Registry:   "ghcr.io/canonical/k8s-snap",
			Repository: "envoyproxy/envoy",
			Tag:        "v1.29.2",
		},
	},
}

var FeatureGateway types.Feature = manifest
