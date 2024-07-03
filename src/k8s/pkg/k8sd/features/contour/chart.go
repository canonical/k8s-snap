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

	// contourGatewayProvisionerEnvoyImageRepo represents the image to use for envoy in the gateway.
	contourGatewayProvisionerEnvoyImageRepo = "ghcr.io/canonical/k8s-snap/envoyproxy/envoy"

	// NOTE: The image version is v1.29.2 instead of 1.28.2
	// to follow the upstream configuration for the contour gateway provisioner.
	// contourGatewayProvisionerEnvoyImageTag is the tag to use for for envoy in the gateway.
	contourGatewayProvisionerEnvoyImageTag = "v1.29.2"

	// contourIngressEnvoyImageRepo represents the image to use for the Contour Envoy proxy.
	contourIngressEnvoyImageRepo = "ghcr.io/canonical/k8s-snap/bitnami/envoy"

	// contourIngressEnvoyImageTag is the tag to use for the Contour Envoy proxy image.
	contourIngressEnvoyImageTag = "1.28.2-debian-12-r0"

	// contourIngressContourImageRepo represents the image to use for Contour.
	contourIngressContourImageRepo = "ghcr.io/canonical/k8s-snap/bitnami/contour"

	// contourIngressContourImageTag is the tag to use for the Contour image.
	contourIngressContourImageTag = "1.28.2-debian-12-r4"

	// contourGatewayProvisionerContourImageRepo represents the image to use for the Contour Gateway Provisioner.
	contourGatewayProvisionerContourImageRepo = "ghcr.io/canonical/k8s-snap/projectcontour/contour"

	// contourGatewayProvisionerContourImageTag is the tag to use for the Contour Gateway Provisioner image.
	contourGatewayProvisionerContourImageTag = "v1.28.2"
)
