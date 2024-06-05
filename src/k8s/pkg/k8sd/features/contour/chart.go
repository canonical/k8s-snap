package contour

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chartContour represents manifests to deploy Contour.
	chartContour = helm.InstallableChart{
		Name:         "ck-ingress",
		Namespace:    "projectcontour",
		ManifestPath: path.Join("charts", "contour-18.1.2.tgz"),
	}

	chartEnvoyGateway = helm.InstallableChart{
		Name:         "ck-gateway",
		Namespace:    "envoy-gateway-system",
		ManifestPath: path.Join("charts", "gateway-helm-v1.0.1.tgz"),
	}
)
