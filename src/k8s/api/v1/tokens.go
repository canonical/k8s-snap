package v1

// WorkerNodeTokenRequest is used to request a token for joining a node to the cluster.
type TokenRequest struct {
	// If true, a token for joining a worker node is created.
	// If false, a token for joining a control plane node is created.
	Worker bool `json:"worker"`
	// Name of the node that should join.
	// Only required for control plane nodes as all workers share the same token.
	Name string `json:"name"`
}

// TokensResponse is used to return a token for joining nodes in the cluster.
type TokensResponse struct {
	// We want to be able to quickly find the tokens in the code, but have the same
	// JSON response for control-plane and worker nodes, thus the discrepancy in naming.
	EncodedToken string `json:"token"`
}
