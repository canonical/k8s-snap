package apiv1

// GetJoinTokenRequest is used to request a token for joining a node to the cluster.
type GetJoinTokenRequest struct {
	// If true, a token for joining a worker node is created.
	// If false, a token for joining a control plane node is created.
	Worker bool `json:"worker"`
	// Name of the node that should join.
	Name string `json:"name"`
}

// GetJoinTokenResponse is used to return a token for joining nodes in the cluster.
type GetJoinTokenResponse struct {
	// We want to be able to quickly find the tokens in the code, but have the same
	// JSON response for control-plane and worker nodes, thus the discrepancy in naming.
	EncodedToken string `json:"token"`
}
