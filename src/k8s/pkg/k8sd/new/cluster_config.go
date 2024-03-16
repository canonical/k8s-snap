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
