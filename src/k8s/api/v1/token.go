package v1

// CreateJoinTokenRequest is used to request a new token to join the cluster.
type CreateJoinTokenRequest struct {
	Name string `json:"name"`
}

// CreateJoinTokenResponse is the response for "POST 1.0/k8sd/tokens".
type CreateJoinTokenResponse struct {
	Token string `json:"token"`
}
