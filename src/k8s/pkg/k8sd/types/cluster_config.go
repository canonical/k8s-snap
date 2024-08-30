package types

type ClusterConfig struct {
	Certificates Certificates `json:"certificates,omitempty"`
	Datastore    Datastore    `json:"datastore,omitempty"`
	APIServer    APIServer    `json:"apiserver,omitempty"`
	Kubelet      Kubelet      `json:"kubelet,omitempty"`

	Network       Network       `json:"network,omitempty"`
	DNS           DNS           `json:"dns,omitempty"`
	Ingress       Ingress       `json:"ingress,omitempty"`
	LoadBalancer  LoadBalancer  `json:"load-balancer,omitempty"`
	Gateway       Gateway       `json:"gateway,omitempty"`
	LocalStorage  LocalStorage  `json:"local-storage,omitempty"`
	MetricsServer MetricsServer `json:"metrics-server,omitempty"`

	Annotations Annotations `json:"annotations,omitempty"`
}
