package v1

// WorkerNodeJoinRequest is used to request a worker node token.
type WorkerNodeJoinRequest struct {
	// Hostname is the name of the worker node.
	Hostname string `json:"name"`
}

// WorkerNodeJoinResponse is used to return a worker node token.
type WorkerNodeJoinResponse struct {
	// EncodedToken is the worker token in encoded form.
	EncodedToken string `json:"token"`
}
