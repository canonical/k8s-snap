package v1

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
