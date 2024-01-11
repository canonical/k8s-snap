// Package api provides the REST API endpoints.
package api

import (
	"github.com/canonical/microcluster/rest"
)

var Endpoints = []rest.Endpoint{
	k8sdCluster,
	k8sdClusterNode,
	k8sdClusterConfig,
	k8sdComponents,
	k8sdComponentsName,
	k8sdToken,

	kubernetesAuthTokens,
	kubernetesAuthWebhook,
}
