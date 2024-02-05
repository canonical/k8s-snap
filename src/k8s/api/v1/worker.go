package v1

// WorkerNodeTokenRequest is used to request a token for joining the cluster as a worker node.
type WorkerNodeTokenRequest struct{}

// WorkerNodeTokenResponse is used to return a token for joining worker nodes in the cluster.
type WorkerNodeTokenResponse struct {
	// We want to be able to quickly find the worker tokens in the code, but have the same
	// JSON response for control-plane and worker nodes, thus the discrepancy in naming.
	EncodedToken string `json:"token"`
}

// WorkerNodeInfoRequest is used by a worker node to retrieve the required credentials
// to join a cluster.
type WorkerNodeInfoRequest struct {
	// Hostname is the name of the worker node.
	Hostname string `json:"hostname"`
}

// WorkerNodeInfoResponse is used to return a worker node token.
type WorkerNodeInfoResponse struct {
	// CA is the PEM encoded certificate authority of the cluster.
	CA string `json:"ca,omitempty"`
	// APIServers is a list of kube-apiserver endpoints of the cluster.
	APIServers []string `json:"apiServers"`
	// KubeletToken is the token to use for kubelet.
	KubeletToken string `json:"kubeletToken"`
	// KubeProxyToken is the token to use for kube-proxy.
	KubeProxyToken string `json:"kubeProxyToken"`
	// ClusterCIDR is the configured cluster CIDR.
	ClusterCIDR string `json:"clusterCIDR"`
	// ClusterDNS is the DNS server address of the cluster.
	ClusterDNS string `json:"clusterDNS,omitempty"`
	// ClusterDomain is the DNS domain of the cluster.
	ClusterDomain string `json:"clusterDomain,omitempty"`
	// CloudProvider is the cloud provider used in the cluster.
	CloudProvider string `json:"cloudProvider,omitempty"`
}
