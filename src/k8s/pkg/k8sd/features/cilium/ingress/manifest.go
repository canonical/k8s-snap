package ingress

import (
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var FeatureIngress types.Feature = types.FeatureManifest{
	Name:    "ingress",
	Version: "1.0.0",

	DefaultValues: map[string]map[string]any{
		cilium_network.CiliumChartName: {
			"ingressController": map[string]any{
				"loadbalancerMode":       "shared",
				"defaultSecretNamespace": "kube-system",
			},
		},
	},
}
