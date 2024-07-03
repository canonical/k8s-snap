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
	// envoyImageRepo represents the image to use for the Contour Envoy proxy.
	envoyImageRepo = "docker.io/bitnami/envoy"

	// envoyImageTag is the tag to use for the Contour Envoy proxy image.
	envoyImageTag = "1.28.2-debian-12-r0"

	// contourImageRepo represents the image to use for Contour.
	contourImageRepo = "docker.io/bitnami/contour"

	// contourImageTag is the tag to use for the Contour image.
	contourImageTag = "1.28.2-debian-12-r4"

	// contourGatewayImageRepo represents the image to use for the Contour Gateway Provisioner.
	contourGatewayImageRepo = "ghcr.io/projectcontour/contour"

	// contourGatewayImageTag is the tag to use for the Contour Gateway Provisioner image.
	contourGatewayImageTag = "v1.28.2"
)
