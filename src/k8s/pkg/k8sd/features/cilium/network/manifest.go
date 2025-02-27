package network

import (
	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

var (
	CiliumChartName = "cilium"

	CiliumAgentImageName    = "cilium-agent"
	CiliumOperatorImageName = "cilium-operator"
)

var FeatureNetwork types.Feature = types.FeatureManifest{
	Name:    "network",
	Version: "1.0.0",
	Charts: map[string]helm.InstallableChart{
		CiliumChartName: {
			Name:             "cilium",
			Version:          "1.16.3",
			InstallName:      "ck-network",
			InstallNamespace: "kube-system",
		},
	},

	Images: map[string]types.Image{
		"cilium-agent": {
			Registry:   "ghcr.io/canonical",
			Repository: "cilium",
			Tag:        "1.16.3-ck0",
		},
		"cilium-operator": {
			Registry:   "ghcr.io/canonical",
			Repository: "cilium-operator",
			Tag:        "1.16.3-ck0",
		},
	},

	DefaultValues: map[string]map[string]any{
		"cilium": {
			"image": map[string]any{
				"useDigest": false,
			},
			"socketLB": map[string]any{
				"enabled": true,
			},
			"cni": map[string]any{
				"confPath": "/etc/cni/net.d",
				"binPath":  "/opt/cni/bin",
			},
			"operator": map[string]any{
				"replicas": 1,
				"image": map[string]any{
					"useDigest": false,
				},
			},
			"envoy": map[string]any{
				"enabled": false, // 1.16+ installs envoy as a standalone daemonset by default if not explicitly disabled
			},
			// https://docs.cilium.io/en/v1.15/network/kubernetes/kubeproxy-free/#kube-proxy-hybrid-modes
			"nodePort": map[string]any{
				"enabled":           true,
				"enableHealthCheck": false,
			},
			"disableEnvoyVersionCheck": true,
			// This flag enables the runtime device detection which is set to true by default in Cilium 1.16+
			"enableRuntimeDeviceDetection": true,
		},
	},
}
