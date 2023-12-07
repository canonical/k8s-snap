package api

import "github.com/canonical/microcluster/rest"

var Endpoints = []rest.Endpoint{
	kubernetesAuthTokens,
	kubernetesAuthWebhook,
}
