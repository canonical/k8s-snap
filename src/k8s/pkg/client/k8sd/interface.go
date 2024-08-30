package k8sd

import (
	"context"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

// ClusterClient implements methods for managing the cluster members.
type ClusterClient interface {
	// BootstrapCluster initializes a new cluster using the provided configuration.
	BootstrapCluster(context.Context, apiv1.BootstrapClusterRequest) (apiv1.BootstrapClusterResponse, error)
	// GetJoinToken generates a token for nodes to join the cluster.
	GetJoinToken(context.Context, apiv1.GetJoinTokenRequest) (apiv1.GetJoinTokenResponse, error)
	// JoinCluster joins an existing cluster.
	JoinCluster(context.Context, apiv1.JoinClusterRequest) error
	// RemoveNode removes a node from the cluster.
	RemoveNode(context.Context, apiv1.RemoveNodeRequest) error
}

// StatusClient implements methods for retrieving the current status of the cluster.
type StatusClient interface {
	// NodeStatus retrieves the current status of the local node.
	// The second return value is false if the node is not part of a cluster.
	NodeStatus(ctx context.Context) (apiv1.NodeStatusResponse, bool, error)
	// ClusterStatus retrieves the current status of the Kubernetes cluster.
	ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatusResponse, error)
}

// ConfigClient implements methods to retrieve and manage the cluster configuration.
type ConfigClient interface {
	// GetClusterConfig retrieves the k8sd cluster configuration.
	GetClusterConfig(context.Context) (apiv1.GetClusterConfigResponse, error)
	// SetClusterConfig updates the k8sd cluster configuration.
	SetClusterConfig(context.Context, apiv1.SetClusterConfigRequest) error
}

// ClusterMaintenanceClient implements methods to manage the cluster.
type ClusterMaintenanceClient interface {
	// RefreshCertificatesPlan generates a plan to refresh the Kubernetes certificates of the node.
	RefreshCertificatesPlan(context.Context, apiv1.RefreshCertificatesPlanRequest) (apiv1.RefreshCertificatesPlanResponse, error)
	// RefreshCertificatesRun refreshes the Kubernetes certificates of the node.
	RefreshCertificatesRun(context.Context, apiv1.RefreshCertificatesRunRequest) (apiv1.RefreshCertificatesRunResponse, error)
}

// UserClient implements methods to enable accessing the cluster.
type UserClient interface {
	// KubeConfig retrieves a kubeconfig file that can be used to access the cluster.
	KubeConfig(context.Context, apiv1.KubeConfigRequest) (apiv1.KubeConfigResponse, error)
}

// ClusterAPIClient implements methods related to ClusterAPI endpoints.
type ClusterAPIClient interface {
	// SetClusterAPIAuthToken sets the well-known token that can be used authenticating requests to the ClusterAPI related endpoints.
	SetClusterAPIAuthToken(context.Context, apiv1.ClusterAPISetAuthTokenRequest) error
}

type Client interface {
	ClusterClient
	StatusClient
	ConfigClient
	ClusterMaintenanceClient
	UserClient
	ClusterAPIClient
}
