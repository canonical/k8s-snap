package v1

// WorkerNodeTokenRequest is used to request a token for joining the cluster as a worker node.
type WorkerNodeTokenRequest struct{}

// WorkerNodeTokenResponse is used to return a token for joining worker nodes in the cluster.
type WorkerNodeTokenResponse struct {
	EncodedToken string `json:"token"`
}

// WorkerNodeInfoRequest is used by a worker node to retrieve the required credentials
// to join a cluster.
type WorkerNodeInfoRequest struct {
	// Hostname is the name of the worker node.
	Hostname string `json:"name"`
}

// WorkerNodeInfoResponse is used to return a worker node token.
type WorkerNodeInfoResponse struct {
	CA             string   `json:"ca,omitempty"`
	APIServers     []string `json:"servers"`
	KubeletToken   string   `json:"kubeletToken"`
	KubeProxyToken string   `json:"proxyToken"`
	ClusterCIDR    string   `json:"clusterCIDR"`
	ClusterDNS     string   `json:"clusterDNS,omitempty"`
	ClusterDomain  string   `json:"clusterDomain,omitempty"`
	CloudProvider  string   `json:"cloudProvider,omitempty"`
}
