package types

type ClusterConfig struct {
	Certificates Certificates `json:"certificates,omitempty"`
	Datastore    Datastore    `json:"datastore,omitempty"`
	APIServer    APIServer    `json:"apiserver,omitempty"`
	Kubelet      Kubelet      `json:"kubelet,omitempty"`
	Containerd   Containerd   `json:"containerd,omitempty"`

	Network       Network       `json:"network,omitempty"`
	DNS           DNS           `json:"dns,omitempty"`
	Ingress       Ingress       `json:"ingress,omitempty"`
	LoadBalancer  LoadBalancer  `json:"load-balancer,omitempty"`
	Gateway       Gateway       `json:"gateway,omitempty"`
	LocalStorage  LocalStorage  `json:"local-storage,omitempty"`
	MetricsServer MetricsServer `json:"metrics-server,omitempty"`
}

func (c ClusterConfig) Empty() bool {
	return c.Certificates.Empty() && c.Datastore.Empty() && c.Network.Empty() && c.APIServer.Empty() && c.Kubelet.Empty() && c.Network.Empty() && c.DNS.Empty() && c.Ingress.Empty() && c.LoadBalancer.Empty() && c.Gateway.Empty() && c.LocalStorage.Empty() && c.MetricsServer.Empty()
}
