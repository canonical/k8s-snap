package metallb

import (
	"path/filepath"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// ChartMetalLB represents manifests to deploy MetalLB speaker and controller.
	ChartMetalLB = helm.InstallableChart{
		Name:         "metallb",
		Namespace:    "metallb-system",
		ManifestPath: filepath.Join("charts", "metallb-0.14.8.tgz"),
	}

	// ChartMetalLBLoadBalancer represents manifests to deploy MetalLB L2 or BGP resources.
	ChartMetalLBLoadBalancer = helm.InstallableChart{
		Name:         "metallb-loadbalancer",
		Namespace:    "metallb-system",
		ManifestPath: filepath.Join("charts", "ck-loadbalancer"),
	}

	// controllerImageRepo is the image to use for metallb-controller.
	controllerImageRepo = "ghcr.io/canonical/metallb-controller"

	// ControllerImageTag is the tag to use for metallb-controller.
	ControllerImageTag = "v0.14.8-ck0"

	// speakerImageRepo is the image to use for metallb-speaker.
	speakerImageRepo = "ghcr.io/canonical/metallb-speaker"

	// speakerImageTag is the tag to use for metallb-speaker.
	speakerImageTag = "v0.14.8-ck0"

	// frrImageRepo is the image to use for frrouting.
	frrImageRepo = "ghcr.io/canonical/k8s-snap/frrouting/frr"

	// frrImageTag is the tag to use for frrouting.
	frrImageTag = "9.0.2"
)
