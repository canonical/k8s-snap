package client

import (
	"context"

	api "github.com/canonical/k8s/api/v1"
	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/snap"
)

// Client defines the interface for interacting with a k8s cluster.
type Client interface {
	// Bootstrap initializes a new cluster member using the provided bootstrap configuration.
	Bootstrap(ctx context.Context, bootstrapConfig apiv1.BootstrapConfig) (apiv1.NodeStatus, error)
	// IsKubernetesAPIServerReady checks if kube-apiserver is reachable.
	IsKubernetesAPIServerReady(ctx context.Context) bool
	// IsBootstrapped checks whether the current node is already bootstrapped.
	IsBootstrapped(ctx context.Context) bool
	// CleanupNode performs cleanup operations for a specific node in the cluster.
	CleanupNode(ctx context.Context, snap snap.Snap, nodeName string)
	// ClusterStatus retrieves the current status of the Kubernetes cluster.
	ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error)
	// NodeStatus retrieves the current status of the local node.
	NodeStatus(ctx context.Context) (apiv1.NodeStatus, error)
	// CreateJoinToken generates a token for a new node to join the cluster.
	CreateJoinToken(ctx context.Context, name string, worker bool) (string, error)
	// GenerateAuthToken generates an authentication token for a specific user with given groups.
	GenerateAuthToken(ctx context.Context, username string, groups []string) (string, error)
	// RevokeAuthToken revokes an authentication token given a token.
	RevokeAuthToken(ctx context.Context, token string) error
	// JoinCluster adds a new node to the cluster using the provided parameters.
	JoinCluster(ctx context.Context, name string, address string, token string) error
	// KubeConfig retrieves the Kubernetes configuration for the current node.
	KubeConfig(ctx context.Context) (string, error)
	// ListComponents returns a list of components in the cluster.
	ListComponents(ctx context.Context) ([]api.Component, error)
	// RemoveNode removes a node from the cluster.
	RemoveNode(ctx context.Context, name string, force bool) error
	// UpdateClusterConfig updates configuration of the cluster.
	UpdateClusterConfig(ctx context.Context, request api.UpdateClusterConfigRequest) error
	// GetClusterConfig retrieves configuration of the cluster.
	GetClusterConfig(ctx context.Context, request api.GetClusterConfigRequest) (api.UserFacingClusterConfig, error)
}
