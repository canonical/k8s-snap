// Package api provides the REST API endpoints.
package api

import (
	"context"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/microcluster/v2/rest"
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
			Path:              "k8sd/cluster",
			Get:               rest.EndpointAction{Handler: e.getClusterStatus, AccessHandler: e.restrictWorkers},
			Post:              rest.EndpointAction{Handler: e.postClusterBootstrap},
			AllowedBeforeInit: true,
		},
		// Node
		// Returns the status (e.g. current role) of the local node (control-plane, worker or unknown).
		{
			Name: "NodeStatus",
			Path: "k8sd/node",
			Get:  rest.EndpointAction{Handler: e.getNodeStatus},
		},
		// Clustering
		// Unified token endpoint for both, control-plane and worker-node.
		{
			Name: "ClusterJoinTokens",
			Path: "k8sd/cluster/tokens",
			Post: rest.EndpointAction{Handler: e.postClusterJoinTokens, AccessHandler: e.restrictWorkers},
		},
		{
			Name: "ClusterJoin",
			Path: "k8sd/cluster/join",
			Post: rest.EndpointAction{Handler: e.postClusterJoin},
			// Joining a node is a bootstrapping action which needs to be available before k8sd is initialized.
			AllowedBeforeInit: true,
		},
		// Cluster removal (control-plane and worker nodes)
		{
			Name: "ClusterRemove",
			Path: "k8sd/cluster/remove",
			Post: rest.EndpointAction{Handler: e.postClusterRemove, AccessHandler: e.restrictWorkers},
		},
		// Worker nodes
		{
			Name: "WorkerInfo",
			Path: "k8sd/worker/info",
			// AllowUntrusted disabled the microcluster authorization check. Authorization is done via custom token.
			Post: rest.EndpointAction{
				Handler:        e.postWorkerInfo,
				AllowUntrusted: true,
				AccessHandler:  ValidateWorkerInfoAccessHandler("worker-name", "worker-token"),
			},
		},
		// Kubeconfig
		{
			Name: "Kubeconfig",
			Path: "k8sd/kubeconfig",
			Get:  rest.EndpointAction{Handler: e.getKubeconfig, AccessHandler: e.restrictWorkers},
		},
		// Get and modify the cluster configuration (e.g. to enable/disable features)
		{
			Name: "ClusterConfig",
			Path: "k8sd/cluster/config",
			Put:  rest.EndpointAction{Handler: e.putClusterConfig, AccessHandler: e.restrictWorkers},
			Get:  rest.EndpointAction{Handler: e.getClusterConfig, AccessHandler: e.restrictWorkers},
		},
		// Kubernetes auth tokens and token review webhook for kube-apiserver
		{
			Name:   "KubernetesAuthTokens",
			Path:   "kubernetes/auth/tokens",
			Get:    rest.EndpointAction{Handler: e.getKubernetesAuthTokens, AllowUntrusted: true},
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
			Name: "ClusterAPI/GenerateJoinToken",
			Path: "x/capi/generate-join-token",
			Post: rest.EndpointAction{Handler: e.postClusterJoinTokens, AccessHandler: ValidateCAPIAuthTokenAccessHandler("capi-auth-token"), AllowUntrusted: true},
		},
		{
			Name: "ClusterAPI/SetAuthToken",
			Path: "x/capi/set-auth-token",
			Post: rest.EndpointAction{Handler: e.postSetClusterAPIAuthToken},
		},
		{
			Name: "ClusterAPI/RemoveNode",
			Path: "x/capi/remove-node",
			Post: rest.EndpointAction{Handler: e.postClusterRemove, AccessHandler: ValidateCAPIAuthTokenAccessHandler("capi-auth-token"), AllowUntrusted: true},
		},
	}
}
