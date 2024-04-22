package v1

// WorkerNodeInfoRequest is used by a worker node to retrieve the required credentials
// to join a cluster.
type WorkerNodeInfoRequest struct {
	// Address is the address of the worker node.
	Address string `json:"address"`
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
	// PodCIDR is the configured CIDR for pods in the cluster.
	PodCIDR string `json:"podCIDR"`
	// ServiceCIDR is the configured CIDR for services in the cluster.
	ServiceCIDR string `json:"serviceCIDR"`
	// ClusterDNS is the DNS server address of the cluster.
	ClusterDNS string `json:"clusterDNS,omitempty"`
	// ClusterDomain is the DNS domain of the cluster.
	ClusterDomain string `json:"clusterDomain,omitempty"`
	// CloudProvider is the cloud provider used in the cluster.
	CloudProvider string `json:"cloudProvider,omitempty"`
	// KubeletCert is the certificate to use for kubelet TLS. It will be empty if the cluster is not using self-signed certificates.
	KubeletCert string `json:"kubeletCrt,omitempty"`
	// KubeletKey is the private key to use for kubelet TLS. It will be empty if the cluster is not using self-signed certificates.
	KubeletKey string `json:"kubeletKey,omitempty"`
	// K8sdPublicKey is the public key that can be used to validate authenticity of cluster messages.
	K8sdPublicKey string `json:"k8sdPublicKey,omitempty"`
}
