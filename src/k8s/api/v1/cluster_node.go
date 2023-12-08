package v1

// AddNodeRequest is used to request to add a node to the cluster.
// The token is generated from the 1.0/k8sd/tokens endpoint and encodes the nodes name.
// The node name is encoded in the url 1.0/k8sd/cluster/{node}.
type AddNodeRequest struct {
	Address string `json:"address"`
	Token   string `json:"token"`
}

// AddNodeResponse is the response from "POST 1.0/k8sd/cluster/{node}"
// TODO: Currently empty, but likely we need to add some information later, thus already registering that type here.
type AddNodeResponse struct{}

// RemoveNodeRequest is used to request to remove a node from the cluster.
// The node name is encoded in the url 1.0/k8sd/cluster/{node}.
type RemoveNodeRequest struct {
	Force bool `json:"force"`
}

// RemoveNodeResponse is the response from "DELETE 1.0/k8sd/cluster/{node}"
type RemoveNodeResponse struct{}
