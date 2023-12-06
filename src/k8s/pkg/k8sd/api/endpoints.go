package api

import "github.com/canonical/microcluster/rest"

var Endpoints = []rest.Endpoint{
	k8sdTokensEndpoint,
	k8sdTokensWebhookEndpoint,
}
