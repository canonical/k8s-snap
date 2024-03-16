package newtypes

type ClusterConfig struct {
	Certificates Certificates `json:"certificates,omitempty"`
	Datastore    Datastore    `json:"datastore,omitempty"`
	Network      Network      `json:"network,omitempty"`
	APIServer    APIServer    `json:"apiserver,omitempty"`
	Kubelet      Kubelet      `json:"kubelet,omitempty"`
	Containerd   Containerd   `json:"containerd,omitempty"`
	Features     Features     `json:"features,omitempty"`
}

func (c ClusterConfig) Empty() bool {
	return c.Certificates.Empty() && c.Datastore.Empty() && c.Network.Empty() && c.APIServer.Empty() && c.Kubelet.Empty() && c.Features.Empty()
}
