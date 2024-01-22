// Package api provides the REST API endpoints.
package api

import (
	"github.com/canonical/microcluster/rest"
)

var Endpoints = []rest.Endpoint{
	k8sdCluster,
	k8sdClusterJoin,
	k8sdClusterNode,
	k8sdDNSComponent,
	k8sdNetworkComponent,
	k8sdComponents,
	k8sdConfig,
	k8sdToken,

	kubernetesAuthTokens,
	kubernetesAuthWebhook,
}
