// Package api provides the REST API endpoints.
package api

import (
	"github.com/canonical/microcluster/rest"
)

var Endpoints = []rest.Endpoint{
	k8sdCluster,
	k8sdWorkerToken,
	k8sdDNSComponent,
	k8sdNetworkComponent,
	k8sdStorageComponent,
	k8sdIngressComponent,
	k8sdGatewayComponent,
	k8sdComponents,
	k8sdConfig,

	kubernetesAuthTokens,
	kubernetesAuthWebhook,
}
