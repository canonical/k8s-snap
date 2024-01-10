package v1

// GetClusterStatusRequest is used to request the current status of the cluster.
type GetClusterStatusRequest struct{}

// GetClusterStatusResponse is the response for "GET 1.0/k8sd/cluster".
type GetClusterStatusResponse struct {
	ClusterStatus ClusterStatus `json:"status"`
}

// InitClusterRequest is used to initialize a k8s cluster.
type InitClusterRequest struct{}

// InitClusterResponse is the response for "POST 1.0/k8sd/cluster".
type InitClusterResponse struct{}

// GetKubeConfigRequest is used to ask for the admin kubeconfig
type GetKubeConfigRequest struct{}

// GetKubeConfigResponse is the response for "GET 1.0/k8sd/cluster/config".
type GetKubeConfigResponse struct {
	KubeConfig string `json:"kube_config"`
}

// ClusterMember holds information about a node in the k8s cluster.
type ClusterMember struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Role        string `json:"role"`
	Fingerprint string `json:"fingerprint"`
	Status      string `json:"status"`
}

// ClusterStatus holds information about the cluster, e.g. its current members
type ClusterStatus struct {
	Members    []ClusterMember `json:"members"`
	Components []Component     `json:"components"`
}

// HaClusterFormed returns true if the cluster is in high-availability mode (more than two voter nodes).
func (c ClusterStatus) HaClusterFormed() bool {
	voters := 0
	for _, member := range c.Members {
		if member.Role == "voter" {
			voters++
		}
	}
	return voters > 2
}
