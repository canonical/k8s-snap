package v1

// GetNodeStatusResponse is the response for "GET 1.0/k8sd/node".
type GetNodeStatusResponse struct {
	NodeStatus NodeStatus `json:"status"`
}
