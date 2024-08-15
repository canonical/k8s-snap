package features

import "github.com/canonical/k8s/pkg/k8sd/types"

const (
	DNS           types.FeatureName = "dns"
	Network       types.FeatureName = "network"
	Gateway       types.FeatureName = "gateway"
	Ingress       types.FeatureName = "ingress"
	LoadBalancer  types.FeatureName = "load-balancer"
	LocalStorage  types.FeatureName = "local-storage"
	MetricsServer types.FeatureName = "metrics-server"
)
