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
		ManifestPath: filepath.Join("charts", "metallb-0.14.9.tgz"),
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
	ControllerImageTag = "5f6feb3539aed39e881d4d9e7ac797fb1760ff407a297ef55b99d2cf50f4be7a-amd64"

	// speakerImageRepo is the image to use for metallb-speaker.
	speakerImageRepo = "ghcr.io/canonical/metallb-speaker"

	// speakerImageTag is the tag to use for metallb-speaker.
	speakerImageTag = "3bc7763599138b9c347198c4e9f227c5cdd9a29a73a9f54546d1bfe265edd2f9-amd64"

	// frrImageRepo is the image to use for frrouting.
	frrImageRepo = "ghcr.io/canonical/frr"

	// frrImageTag is the tag to use for frrouting.
	frrImageTag = "ede05e1e85736e61b14d3e2f642c3dcb76f2f91fda49447d9cbade9a6d013f1c-amd64"
)
