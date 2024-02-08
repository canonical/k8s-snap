package v1

// JoinClusterRequest is used to request to add a node to the cluster.
type JoinClusterRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Token   string `json:"token"`
}

// JoinClusterResponse is the response from "POST 1.0/k8sd/cluster/{node}"
type JoinClusterResponse struct{}
