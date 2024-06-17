package contour

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chartContour represents manifests to deploy Contour.
	// This excludes shared CRDs.
	chartContour = helm.InstallableChart{
		Name:         "ck-ingress",
		Namespace:    "projectcontour",
		ManifestPath: path.Join("charts", "contour-17.0.4.tgz"),
	}
	// chartGateway represents manifests to deploy Contour Gateway.
	// This excludes shared CRDs.
	chartGateway = helm.InstallableChart{
		Name:         "ck-gateway",
		Namespace:    "projectcontour",
		ManifestPath: path.Join("charts", "ck-gateway-contour-1.28.2.tgz"),
	}
	// chartDefaultTLS represents manifests to deploy a delegation resource for the default TLS secret.
	chartDefaultTLS = helm.InstallableChart{
		Name:         "ck-ingress-tls",
		Namespace:    "projectcontour-root",
		ManifestPath: path.Join("charts", "ck-ingress-tls"),
	}
	// chartCommonContourCRDS represents manifests to deploy common Contour CRDs.
	chartCommonContourCRDS = helm.InstallableChart{
		Name:         "ck-contour-common",
		Namespace:    "projectcontour",
		ManifestPath: path.Join("charts", "ck-contour-common-1.28.2.tgz"),
	}
)
