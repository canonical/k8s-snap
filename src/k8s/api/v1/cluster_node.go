package v1

// JoinNodeRequest is used to request to add a node to the cluster.
type JoinNodeRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Token   string `json:"token"`
}

// JoinNodeResponse is the response from "POST 1.0/k8sd/cluster/{node}"
type JoinNodeResponse struct{}
