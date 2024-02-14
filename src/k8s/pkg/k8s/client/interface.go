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
	Bootstrap(ctx context.Context, bootstrapConfig apiv1.BootstrapConfig) (apiv1.ClusterMember, error)
	// IsBootstrapped checks whether the current node is already bootstrapped.
	IsBootstrapped(ctx context.Context) bool
	// CleanupNode performs cleanup operations for a specific node in the cluster.
	CleanupNode(ctx context.Context, snap snap.Snap, nodeName string)
	// ClusterStatus retrieves the current status of the Kubernetes cluster.
	ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error)
	// CreateJoinToken generates a token for a new node to join the cluster.
	CreateJoinToken(ctx context.Context, name string, worker bool) (string, error)
	// GenerateAuthToken generates an authentication token for a specific user with given groups.
	GenerateAuthToken(ctx context.Context, username string, groups []string) (string, error)
	// JoinCluster adds a new node to the cluster using the provided parameters.
	JoinCluster(ctx context.Context, name string, address string, token string) error
	// KubeConfig retrieves the Kubernetes configuration for the current node.
	KubeConfig(ctx context.Context) (string, error)
	// ListComponents returns a list of components in the cluster.
	ListComponents(ctx context.Context) ([]api.Component, error)
	// RemoveNode removes a node from the cluster.
	RemoveNode(ctx context.Context, name string, force bool) error
	// UpdateDNSComponent updates the DNS component in the cluster.
	UpdateDNSComponent(ctx context.Context, request api.UpdateDNSComponentRequest) error
	// UpdateGatewayComponent updates the Gateway component in the cluster.
	UpdateGatewayComponent(ctx context.Context, request api.UpdateGatewayComponentRequest) error
	// UpdateIngressComponent updates the Ingress component in the cluster.
	UpdateIngressComponent(ctx context.Context, request api.UpdateIngressComponentRequest) error
	// UpdateLoadBalancerComponent updates the Load Balancer component in the cluster.
	UpdateLoadBalancerComponent(ctx context.Context, request api.UpdateLoadBalancerComponentRequest) error
	// UpdateNetworkComponent updates the Network component in the cluster.
	UpdateNetworkComponent(ctx context.Context, request api.UpdateNetworkComponentRequest) error
	// UpdateStorageComponent updates the Storage component in the cluster.
	UpdateStorageComponent(ctx context.Context, request api.UpdateStorageComponentRequest) error
}
