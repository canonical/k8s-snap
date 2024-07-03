package apiv1

// SetClusterAPIAuthTokenRequest is used to request to set the auth token for ClusterAPI.
type SetClusterAPIAuthTokenRequest struct {
	Token string `json:"token"`
}
