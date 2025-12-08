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
		ManifestPath: filepath.Join("charts", "metallb-0.15.3.tgz"),
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
	ControllerImageTag = "v0.15.3-ck0"

	// speakerImageRepo is the image to use for metallb-speaker.
	speakerImageRepo = "ghcr.io/canonical/metallb-speaker"

	// speakerImageTag is the tag to use for metallb-speaker.
	speakerImageTag = "v0.15.3-ck0"

	// frrImageRepo is the image to use for frrouting.
	frrImageRepo = "ghcr.io/canonical/frr"

	// frrImageTag is the tag to use for frrouting.
	frrImageTag = "10.4.1-ck0"
)
