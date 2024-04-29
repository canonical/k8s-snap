package client

import (
	"context"

	apiv1 "github.com/canonical/k8s/api/v1"
)

// Client defines the interface for interacting with a k8s cluster.
type Client interface {
	// Bootstrap initializes a new cluster member using the provided bootstrap configuration.
	Bootstrap(ctx context.Context, request apiv1.PostClusterBootstrapRequest) (apiv1.NodeStatus, error)
	// IsBootstrapped checks whether the current node is already bootstrapped.
	IsBootstrapped(ctx context.Context) bool
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
	// DeleteClusterMember removes a node from the cluster.
	DeleteClusterMember(ctx context.Context, request apiv1.RemoveNodeRequest) error
	// ResetClusterMember calls microcluster's ResetClusterMember.
	ResetClusterMember(ctx context.Context, name string, force bool) error
	// CleanupKubernetesServices cleans up service arguments and leftover state on the local node, typically used after a failed bootstrap or join attempt.
	CleanupKubernetesServices(ctx context.Context) error
	// UpdateClusterConfig updates configuration of the cluster.
	UpdateClusterConfig(ctx context.Context, request apiv1.UpdateClusterConfigRequest) error
	// GetClusterConfig retrieves configuration of the cluster.
	GetClusterConfig(ctx context.Context, request apiv1.GetClusterConfigRequest) (apiv1.UserFacingClusterConfig, error)
	// WaitForMicroclusterNodeToBeReady waits until the underlying dqlite node of the microcluster is not in PENDING state.
	WaitForMicroclusterNodeToBeReady(ctx context.Context, nodeName string) error
}

var _ Client = &k8sdClient{}
