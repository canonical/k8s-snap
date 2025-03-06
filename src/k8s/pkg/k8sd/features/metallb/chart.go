package metallb

import (
	"embed"

	"github.com/canonical/k8s/pkg/client/helm"
)

//go:embed all:charts
var ChartFS embed.FS

var (
	// ChartMetalLB represents manifests to deploy MetalLB speaker and controller.
	ChartMetalLB = helm.InstallableChart{
		Name:             "metallb",
		Version:          "0.14.9",
		InstallName:      "metallb",
		InstallNamespace: "metallb-system",
	}

	// ChartMetalLBLoadBalancer represents manifests to deploy MetalLB L2 or BGP resources.
	ChartMetalLBLoadBalancer = helm.InstallableChart{
		Name:             "ck-loadbalancer",
		Version:          "0.1.1",
		InstallName:      "metallb-loadbalancer",
		InstallNamespace: "metallb-system",
	}

	// controllerImageRepo is the image to use for metallb-controller.
	controllerImageRepo = "ghcr.io/canonical/metallb-controller"

	// ControllerImageTag is the tag to use for metallb-controller.
	ControllerImageTag = "v0.14.9-ck0"

	// speakerImageRepo is the image to use for metallb-speaker.
	speakerImageRepo = "ghcr.io/canonical/metallb-speaker"

	// speakerImageTag is the tag to use for metallb-speaker.
	speakerImageTag = "v0.14.9-ck0"

	// frrImageRepo is the image to use for frrouting.
	frrImageRepo = "ghcr.io/canonical/frr"

	// frrImageTag is the tag to use for frrouting.
	frrImageTag = "9.1.3"
)
