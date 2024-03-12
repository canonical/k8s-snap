// Package api provides the REST API endpoints.
package api

import (
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/rest"
)

// Endpoints returns the list of endpoints for a given microcluster app.
func Endpoints(app *microcluster.MicroCluster) []rest.Endpoint {
	return []rest.Endpoint{
		// Cluster status
		{
			Name: "ClusterStatus",
			Path: "k8sd/cluster",
			Get:  rest.EndpointAction{Handler: getClusterStatus, AccessHandler: RestrictWorkers},
		},
		// Node
		// Returns the status (e.g. current role) of the local node (control-plane, worker or unknown).
		{
			Name: "NodeStatus",
			Path: "k8sd/node",
			Get:  rest.EndpointAction{Handler: getNodeStatus},
		},
		// Clustering
		// Unified token endpoint for both, control-plane and worker-node.
		{
			Name: "ClusterJoinTokens",
			Path: "k8sd/cluster/tokens",
			Post: rest.EndpointAction{Handler: wrapHandlerWithMicroCluster(app, postClusterJoinTokens), AccessHandler: RestrictWorkers},
		},
		{
			Name: "ClusterJoin",
			Path: "k8sd/cluster/join",
			Post: rest.EndpointAction{Handler: wrapHandlerWithMicroCluster(app, postClusterJoin)},
			// Joining a node is a bootstrapping action which needs to be available before k8sd is initialized.
			AllowedBeforeInit: true,
		},
		// Cluster removal (control-plane and worker nodes)
		{
			Name: "ClusterRemove",
			Path: "k8sd/cluster/remove",
			Post: rest.EndpointAction{Handler: wrapHandlerWithMicroCluster(app, postClusterRemove), AccessHandler: RestrictWorkers},
		},
		// Worker nodes
		{
			Name: "WorkerInfo",
			Path: "k8sd/worker/info",
			// AllowUntrusted disabled the microcluster authorization check. Authorization is done via custom token.
			Post: rest.EndpointAction{Handler: postWorkerInfo, AllowUntrusted: true, AccessHandler: TokenAuthentication},
		},
		// Kubeconfig
		{
			Name: "Kubeconfig",
			Path: "k8sd/kubeconfig",
			Get:  rest.EndpointAction{Handler: getKubeconfig, AccessHandler: RestrictWorkers},
		},
		// Get and modify the cluster configuration (e.g. to enable/disable functionalities)
		{
			Name: "ClusterConfig",
			Path: "k8sd/cluster/config",
			Put:  rest.EndpointAction{Handler: putClusterConfig, AccessHandler: RestrictWorkers},
			Get:  rest.EndpointAction{Handler: getClusterConfig, AccessHandler: RestrictWorkers},
		},
		// Kubernetes auth tokens and token review webhook for kube-apiserver
		{
			Name:   "KubernetesAuthTokens",
			Path:   "kubernetes/auth/tokens",
			Get:    rest.EndpointAction{Handler: getKubernetesAuthTokens, AllowUntrusted: true},
			Post:   rest.EndpointAction{Handler: postKubernetesAuthTokens},
			Delete: rest.EndpointAction{Handler: deleteKubernetesAuthTokens},
		},
		{
			Name: "KubernetesAuthWebhook",
			Path: "kubernetes/auth/webhook",
			Post: rest.EndpointAction{Handler: postKubernetesAuthWebhook, AllowUntrusted: true},
		},
	}
}
