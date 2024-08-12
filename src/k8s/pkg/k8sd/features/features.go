package features

type FeatureName string

const (
	DNS           FeatureName = "dns"
	Network       FeatureName = "network"
	Gateway       FeatureName = "gateway"
	Ingress       FeatureName = "ingress"
	LoadBalancer  FeatureName = "load-balancer"
	LocalStorage  FeatureName = "local-storage"
	MetricsServer FeatureName = "metrics-server"
)
