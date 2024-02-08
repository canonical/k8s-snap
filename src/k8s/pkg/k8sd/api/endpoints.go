// Package api provides the REST API endpoints.
package api

import (
	"github.com/canonical/microcluster/rest"
)

var Endpoints = []rest.Endpoint{
	// Cluster status
	{
		Name: "ClusterStatus",
		Path: "k8sd/cluster",
		Get:  rest.EndpointAction{Handler: getClusterStatus, AccessHandler: WorkerRestricted},
	},
	// Clustering
	// Unified token endpoint for both, control-plane and worker-node.
	{
		Name: "Tokens",
		Path: "k8sd/cluster/tokens",
		Post: rest.EndpointAction{Handler: postTokens},
	},
	{
		Name: "JoinNode",
		Path: "k8sd/cluster/join",
		Post: rest.EndpointAction{Handler: postClusterNode},
		// Joining a node is a bootstrapping action which needs to be available before k8sd is initialized.
		AllowedBeforeInit: true,
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
		Get:  rest.EndpointAction{Handler: getKubeconfig},
	},
	// Cluster components
	{
		Name: "Components",
		Path: "k8sd/components",
		Get:  rest.EndpointAction{Handler: getComponents},
	},
	{
		Name: "DNSComponent",
		Path: "k8sd/components/dns",
		Put:  rest.EndpointAction{Handler: putDNSComponent},
	},
	{
		Name: "NetworkComponent",
		Path: "k8sd/components/network",
		Put:  rest.EndpointAction{Handler: putNetworkComponent},
	},
	{
		Name: "StorageComponent",
		Path: "k8sd/components/storage",
		Put:  rest.EndpointAction{Handler: putStorageComponent},
	},
	{
		Name: "IngressComponent",
		Path: "k8sd/components/ingress",
		Put:  rest.EndpointAction{Handler: putIngressComponent},
	},
	{
		Name: "GatewayComponent",
		Path: "k8sd/components/gateway",
		Put:  rest.EndpointAction{Handler: putGatewayComponent},
	},
	{
		Name: "LoadBalancerComponent",
		Path: "k8sd/components/loadbalancer",
		Put:  rest.EndpointAction{Handler: putLoadBalancerComponent},
	},
	// Kubernetes auth tokens and token review webhook for kube-apiserver
	{
		Name: "KubernetesAuthTokens",
		Path: "kubernetes/auth/tokens",
		Get:  rest.EndpointAction{Handler: getKubernetesAuthToken, AllowUntrusted: true},
		Post: rest.EndpointAction{Handler: postKubernetesAuthToken},
	},
	{
		Name: "KubernetesAuthWebhook",
		Path: "kubernetes/auth/webhook",
		Post: rest.EndpointAction{Handler: postKubernetesAuthWebhook, AllowUntrusted: true},
	},
}
