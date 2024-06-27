package metallb

import (
	"path"

	"github.com/canonical/k8s/pkg/client/helm"
)

var (
	// chartMetalLB represents manifests to deploy MetalLB speaker and controller.
	chartMetalLB = helm.InstallableChart{
		Name:         "metallb",
		Namespace:    "metallb-system",
		ManifestPath: path.Join("charts", "metallb-0.14.5.tgz"),
	}

	// chartMetalLBLoadBalancer represents manifests to deploy MetalLB L2 or BGP resources.
	chartMetalLBLoadBalancer = helm.InstallableChart{
		Name:         "metallb-loadbalancer",
		Namespace:    "metallb-system",
		ManifestPath: path.Join("charts", "ck-loadbalancer"),
	}

	// controllerImageRepo is the image to use for metallb-controller.
	controllerImageRepo = "ghcr.io/canonical/k8s-snap/metallb/controller"

	// controllerImageTag is the tag to use for metallb-controller.
	controllerImageTag = "v0.14.5"

	// speakerImageRepo is the image to use for metallb-speaker.
	speakerImageRepo = "ghcr.io/canonical/k8s-snap/metallb/speaker"

	// speakerImageTag is the tag to use for metallb-speaker.
	speakerImageTag = "v0.14.5"

	// frrImageRepo is the image to use for frrouting.
	frrImageRepo = "ghcr.io/canonical/k8s-snap/frrouting/frr"

	// frrImageTag is the tag to use for frrouting.
	frrImageTag = "9.0.2"
)
