package api

import "github.com/canonical/microcluster/rest"

var Endpoints = []rest.Endpoint{
	{
		Name: "K8sdTokens",
		Path: "k8sd/tokens",
		Get:  rest.EndpointAction{Handler: getK8sdToken, AllowUntrusted: true},
		Post: rest.EndpointAction{Handler: postK8sdToken},
	},
	{
		Name: "K8sdTokensWebhook",
		Path: "k8sd/tokens/webhook",
		Post: rest.EndpointAction{Handler: tokenReviewWebhook, AllowUntrusted: true},
	},
}
