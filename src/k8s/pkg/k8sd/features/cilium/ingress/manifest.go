package ingress

import (
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var manifest = types.FeatureManifest{
	Name:    "ingress",
	Version: "1.0.0",
}

var FeatureIngress types.Feature = manifest
