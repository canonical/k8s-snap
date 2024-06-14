package v1

// SetAuthTokenRequest is used to request to set the auth token for ClusterAPI.
type SetAuthTokenRequest struct {
	Token string `json:"token"`
}
