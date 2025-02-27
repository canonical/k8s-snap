package loadbalancer

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var LoadbalancerChartName = "ck-loadbalancer"

var manifest = types.FeatureManifest{
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
}

var FeatureLoadBalancer types.Feature = manifest
