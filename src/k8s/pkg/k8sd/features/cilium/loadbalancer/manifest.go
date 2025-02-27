package loadbalancer

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	LoadbalancerChartName = "ck-loadbalancer"
)

var FeatureLoadBalancer types.Feature = types.FeatureManifest{
	Name:    "loadbalancer",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		LoadbalancerChartName: {
			Name:             "ck-loadbalancer",
			Version:          "0.1.1",
			InstallName:      "ck-loadbalancer",
			InstallNamespace: "kube-system",
		},
	},

	DefaultValues: map[string]map[string]any{
		"cilium": {
			"externalIPs": map[string]any{
				"enabled": true,
			},
			// https://docs.cilium.io/en/v1.14/network/l2-announcements/#sizing-client-rate-limit
			// Assuming for 50 LB services
			"k8sClientRateLimit": map[string]any{
				"qps":   10,
				"burst": 20,
			},
		},
		"ck-loadbalancer": {
			"driver": "cilium",
		},
	},
}
