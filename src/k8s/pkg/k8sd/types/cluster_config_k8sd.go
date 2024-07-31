package types

// K8sd contains configuration around the k8sd lifecycle.
type K8sd struct {
	// ShouldRemoveK8sNode is a flag to indicate whether the k8s node should be removed.
	// If set, the k8s node will be removed. If not set, only the microcluster & file cleanup is done.
	// This is useful, if an external controller (e.g. CAPI) is responsible for the Kubernetes node life cycle.
	ShouldRemoveK8sNode *bool `json:"should-remove-k8s-node,omitempty"`
}

func (c K8sd) GetShouldRemoveK8sNode() bool { return getField(c.ShouldRemoveK8sNode) }
func (c K8sd) Empty() bool                  { return c == K8sd{} }
