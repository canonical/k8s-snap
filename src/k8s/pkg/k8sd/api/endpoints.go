// Package api provides the REST API endpoints.
package api

import (
	"context"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/microcluster/v3/rest"
)

type Endpoints struct {
	context  context.Context
	provider Provider
}

// New creates a new API server instance.
// Context is the context to use for the API servers endpoints.
func New(ctx context.Context, provider Provider) map[string]rest.Server {
	k8sd := &Endpoints{
		context:  ctx,
		provider: provider,
	}
	return map[string]rest.Server{
		"k8sd": {
			CoreAPI:   true,
			ServeUnix: true,
			PreInit:   true,
			Resources: []rest.Resources{
				{
					PathPrefix: apiv1.K8sdAPIVersion,
					Endpoints:  k8sd.Endpoints(),
				},
			},
		},
	}
}

func (e *Endpoints) Context() context.Context {
	return e.context
}

// Endpoints returns the list of endpoints for a given microcluster app.
func (e *Endpoints) Endpoints() []rest.Endpoint {
	return []rest.Endpoint{
		// Cluster status and bootstrap
		{
			Name:              "Cluster",
			Path:              apiv1.BootstrapClusterRPC, // == apiv1.ClusterStatusRPC
			Get:               rest.EndpointAction{Handler: e.getClusterStatus, AccessHandler: e.restrictWorkers},
			Post:              rest.EndpointAction{Handler: e.postClusterBootstrap},
			AllowedBeforeInit: true,
		},
		// Node
		// Returns the status (e.g. current role) of the local node (control-plane, worker or unknown).
		{
			Name: "NodeStatus",
			Path: apiv1.NodeStatusRPC,
			Get:  rest.EndpointAction{Handler: e.getNodeStatus},
		},
		// Clustering
		// Unified token endpoint for both, control-plane and worker-node.
		{
			Name: "GetJoinToken",
			Path: apiv1.GetJoinTokenRPC,
			Post: rest.EndpointAction{Handler: e.postClusterJoinTokens, AccessHandler: e.restrictWorkers},
		},
		{
			Name: "JoinCluster",
			Path: apiv1.JoinClusterRPC,
			Post: rest.EndpointAction{Handler: e.postClusterJoin},
			// Joining a node is a bootstrapping action which needs to be available before k8sd is initialized.
			AllowedBeforeInit: true,
		},
		// Cluster removal (control-plane and worker nodes)
		{
			Name: "RemoveNode",
			Path: apiv1.RemoveNodeRPC,
			Post: rest.EndpointAction{Handler: e.postClusterRemove, AccessHandler: e.restrictWorkers},
		},
		// Worker nodes
		{
			Name: "GetWorkerJoinInfo",
			Path: apiv1.GetWorkerJoinInfoRPC,
			// AllowUntrusted disabled the microcluster authorization check. Authorization is done via custom token.
			Post: rest.EndpointAction{
				Handler:        e.postWorkerInfo,
				AllowUntrusted: true,
				AccessHandler:  ValidateWorkerInfoAccessHandler("worker-name", "worker-token"),
			},
		},
		// Certificates
		{
			Name: "RefreshCerts/Plan",
			Path: apiv1.RefreshCertificatesPlanRPC,
			Post: rest.EndpointAction{Handler: e.postRefreshCertsPlan},
		},
		{
			Name: "RefreshCerts/Run",
			Path: apiv1.RefreshCertificatesRunRPC,
			Post: rest.EndpointAction{Handler: e.postRefreshCertsRun},
		},
		// Kubeconfig
		{
			Name: "Kubeconfig",
			Path: apiv1.KubeConfigRPC,
			Get:  rest.EndpointAction{Handler: e.getKubeconfig, AccessHandler: e.restrictWorkers},
		},
		// Get and modify the cluster configuration (e.g. to enable/disable features)
		{
			Name: "ClusterConfig",
			Path: apiv1.GetClusterConfigRPC, // == apiv1.SetClusterConfigRPC
			Put:  rest.EndpointAction{Handler: e.putClusterConfig, AccessHandler: e.restrictWorkers},
			Get:  rest.EndpointAction{Handler: e.getClusterConfig, AccessHandler: e.restrictWorkers},
		},
		// Kubernetes auth tokens and token review webhook for kube-apiserver
		{
			Name:   "KubernetesAuthTokens",
			Path:   apiv1.GenerateKubernetesAuthTokenRPC, // == apiv1.RevokeKubernetesAuthTokenRPC
			Post:   rest.EndpointAction{Handler: e.postKubernetesAuthTokens},
			Delete: rest.EndpointAction{Handler: e.deleteKubernetesAuthTokens},
		},
		{
			Name: "KubernetesAuthWebhook",
			Path: "kubernetes/auth/webhook",
			Post: rest.EndpointAction{Handler: e.postKubernetesAuthWebhook, AllowUntrusted: true},
		},
		// ClusterAPI management endpoints.
		{
			Name: "ClusterAPI/GetJoinToken",
			Path: apiv1.ClusterAPIGetJoinTokenRPC,
			Post: rest.EndpointAction{Handler: e.postClusterJoinTokens, AccessHandler: ValidateCAPIAuthTokenAccessHandler("capi-auth-token"), AllowUntrusted: true},
		},
		{
			Name: "ClusterAPI/SetAuthToken",
			Path: apiv1.ClusterAPISetAuthTokenRPC,
			Post: rest.EndpointAction{Handler: e.postSetClusterAPIAuthToken},
		},
		{
			Name: "ClusterAPI/RemoveNode",
			Path: apiv1.ClusterAPIRemoveNodeRPC,
			Post: rest.EndpointAction{Handler: e.postClusterRemove, AccessHandler: ValidateCAPIAuthTokenAccessHandler("capi-auth-token"), AllowUntrusted: true},
		},
		{
			Name: "ClusterAPI/CertificatesExpiry",
			Path: apiv1.ClusterAPICertificatesExpiryRPC,
			Post: rest.EndpointAction{Handler: e.postCertificatesExpiry, AccessHandler: e.ValidateNodeTokenAccessHandler("node-token"), AllowUntrusted: true},
		},
		{
			Name: "ClusterAPI/RefreshCerts/Plan",
			Path: apiv1.ClusterAPICertificatesPlanRPC,
			Post: rest.EndpointAction{Handler: e.postRefreshCertsPlan, AccessHandler: e.ValidateNodeTokenAccessHandler("node-token"), AllowUntrusted: true},
		},
		{
			Name: "ClusterAPI/RefreshCerts/Run",
			Path: apiv1.ClusterAPICertificatesRunRPC,
			Post: rest.EndpointAction{Handler: e.postRefreshCertsRun, AccessHandler: e.ValidateNodeTokenAccessHandler("node-token"), AllowUntrusted: true},
		},
		{
			Name: "ClusterAPI/RefreshCerts/Approve",
			Path: "x/capi/refresh-certs/approve",
			Post: rest.EndpointAction{Handler: e.postApproveWorkerCSR, AccessHandler: ValidateCAPIAuthTokenAccessHandler("capi-auth-token"), AllowUntrusted: true},
		},
		// Snap refreshes
		{
			Name: "Snap/Refresh",
			Path: apiv1.SnapRefreshRPC,
			Post: rest.EndpointAction{Handler: e.postSnapRefresh, AccessHandler: e.ValidateNodeTokenAccessHandler("node-token"), AllowUntrusted: true},
		},
		{
			Name: "Snap/RefreshStatus",
			Path: apiv1.SnapRefreshStatusRPC,
			Post: rest.EndpointAction{Handler: e.postSnapRefreshStatus, AccessHandler: e.ValidateNodeTokenAccessHandler("node-token"), AllowUntrusted: true},
		},
	}
}
