package api

import "github.com/canonical/microcluster/rest"

var Endpoints = []rest.Endpoint{
	{
		Name: "KubernetesAuthTokens",
		Path: "kubernetes/auth/tokens",
		Get:  rest.EndpointAction{Handler: postKubernetesAuthToken, AllowUntrusted: true},
		Post: rest.EndpointAction{Handler: getKubernetesAuthToken},
	},
	{
		Name: "KubernetesAuthWebhook",
		Path: "kubernetes/auth/webhook",
		Post: rest.EndpointAction{Handler: kubernetesAuthTokenReviewWebhook, AllowUntrusted: true},
	},
}
