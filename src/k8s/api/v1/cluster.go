package v1

// GetClusterStatusRequest is used to request the current status of the cluster.
type GetClusterStatusRequest struct{}

// GetClusterStatusResponse is the response for "GET 1.0/k8sd/cluster".
type GetClusterStatusResponse struct {
	ClusterStatus ClusterStatus `json:"status"`
}

// PostClusterBootstrapRequest is used to bootstrap the cluster by using the endpoint "POST 1.0/k8sd/cluster".
type PostClusterBootstrapRequest struct {
	Name    string          `json:"name"`
	Address string          `json:"address"`
	Config  BootstrapConfig `json:"config"`
}

// GetKubeConfigRequest is used to ask for the admin kubeconfig
type GetKubeConfigRequest struct {
	Server string `json:"server"`
}

// GetKubeConfigResponse is the response for "GET 1.0/k8sd/cluster/config".
type GetKubeConfigResponse struct {
	KubeConfig string `json:"kubeconfig"`
}
