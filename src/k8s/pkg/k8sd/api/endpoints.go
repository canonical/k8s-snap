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
			Post: rest.EndpointAction{Handler: wrapHandlerWithMicroCluster(app, postClusterRemoveNode), AccessHandler: RestrictWorkers},
		},
		// Worker nodes
		{
			Name: "WorkerInfo",
			Path: "k8sd/worker/info",
			// This endpoint is used by worker nodes that are not part of the microcluster.
			// We authenticate by passing a token through an HTTP header instead.
			Post: rest.EndpointAction{Handler: postWorkerInfo, AllowUntrusted: true},
		},
		// Kubeconfig
		{
			Name: "Kubeconfig",
			Path: "k8sd/kubeconfig",
			Get:  rest.EndpointAction{Handler: getKubeconfig, AccessHandler: RestrictWorkers},
		},
		// Cluster components
		{
			Name: "Components",
			Path: "k8sd/components",
			Get:  rest.EndpointAction{Handler: getComponents, AccessHandler: RestrictWorkers},
		},
		{
			Name: "DNSComponent",
			Path: "k8sd/components/dns",
			Put:  rest.EndpointAction{Handler: putDNSComponent, AccessHandler: RestrictWorkers},
		},
		{
			Name: "NetworkComponent",
			Path: "k8sd/components/network",
			Put:  rest.EndpointAction{Handler: putNetworkComponent, AccessHandler: RestrictWorkers},
		},
		{
			Name: "StorageComponent",
			Path: "k8sd/components/storage",
			Put:  rest.EndpointAction{Handler: putStorageComponent, AccessHandler: RestrictWorkers},
		},
		{
			Name: "IngressComponent",
			Path: "k8sd/components/ingress",
			Put:  rest.EndpointAction{Handler: putIngressComponent, AccessHandler: RestrictWorkers},
		},
		{
			Name: "GatewayComponent",
			Path: "k8sd/components/gateway",
			Put:  rest.EndpointAction{Handler: putGatewayComponent, AccessHandler: RestrictWorkers},
		},
		{
			Name: "LoadBalancerComponent",
			Path: "k8sd/components/loadbalancer",
			Put:  rest.EndpointAction{Handler: putLoadBalancerComponent, AccessHandler: RestrictWorkers},
		},
		{
			Name: "MetricsServerComponent",
			Path: "k8sd/components/metrics-server",
			Put:  rest.EndpointAction{Handler: putMetricsServerComponent, AccessHandler: RestrictWorkers},
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
