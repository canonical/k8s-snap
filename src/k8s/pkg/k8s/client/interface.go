package client

import (
	"context"

	apiv1 "github.com/canonical/k8s/api/v1"
)

// Client defines the interface for interacting with a k8s cluster.
type Client interface {
	// Bootstrap initializes a new cluster member using the provided bootstrap configuration.
	Bootstrap(ctx context.Context, name string, address string, bootstrapConfig apiv1.BootstrapConfig) (apiv1.NodeStatus, error)
	// IsKubernetesAPIServerReady checks if kube-apiserver is reachable.
	IsKubernetesAPIServerReady(ctx context.Context) bool
	// IsBootstrapped checks whether the current node is already bootstrapped.
	IsBootstrapped(ctx context.Context) bool
	// CleanupNode performs cleanup operations for a specific node in the cluster.
	CleanupNode(ctx context.Context, nodeName string)
	// ClusterStatus retrieves the current status of the Kubernetes cluster.
	ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error)
	// LocalNodeStatus retrieves the current status of the local node.
	LocalNodeStatus(ctx context.Context) (apiv1.NodeStatus, error)
	// GetJoinToken generates a token for a new node to join the cluster.
	GetJoinToken(ctx context.Context, request apiv1.GetJoinTokenRequest) (string, error)
	// GenerateAuthToken generates an authentication token for a specific user with given groups.
	GenerateAuthToken(ctx context.Context, request apiv1.GenerateKubernetesAuthTokenRequest) (string, error)
	// RevokeAuthToken revokes an authentication token given a token.
	RevokeAuthToken(ctx context.Context, request apiv1.RevokeKubernetesAuthTokenRequest) error
	// JoinCluster adds a new node to the cluster using the provided parameters.
	JoinCluster(ctx context.Context, request apiv1.JoinClusterRequest) error
	// KubeConfig retrieves the Kubernetes configuration for the current node.
	KubeConfig(ctx context.Context, request apiv1.GetKubeConfigRequest) (string, error)
	// RemoveNode removes a node from the cluster.
	RemoveNode(ctx context.Context, request apiv1.RemoveNodeRequest) error
	// UpdateClusterConfig updates configuration of the cluster.
	UpdateClusterConfig(ctx context.Context, request apiv1.UpdateClusterConfigRequest) error
	// GetClusterConfig retrieves configuration of the cluster.
	GetClusterConfig(ctx context.Context, request apiv1.GetClusterConfigRequest) (apiv1.UserFacingClusterConfig, error)
}

var _ Client = &k8sdClient{}
