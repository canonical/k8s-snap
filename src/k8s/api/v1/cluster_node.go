package apiv1

import "time"

// JoinClusterRequest is used to request to add a node to the cluster.
type JoinClusterRequest struct {
	Name    string        `json:"name"`
	Address string        `json:"address"`
	Token   string        `json:"token"`
	Config  string        `json:"config"`
	Timeout time.Duration `json:"timeout"`
}

// RemoveNodeRequest is used to request to remove a node from the cluster.
type RemoveNodeRequest struct {
	Name  string `json:"name"`
	Force bool   `json:"force"`
}
