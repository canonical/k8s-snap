package kubernetes

import (
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is a wrapper around the kubernetes.Interface.
type Client struct {
	kubernetes.Interface
	ctrlclient.Client
	config *rest.Config
}

func NewClient(restClientGetter genericclioptions.RESTClientGetter) (*Client, error) {
	config, err := restClientGetter.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build Kubernetes REST config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	scheme, err := NewScheme()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheme: %w", err)
	}

	ctrlC, err := ctrlclient.New(config, ctrlclient.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}

	return &Client{
		Interface: clientset,
		Client:    ctrlC,
		config:    config,
	}, nil
}

func (c *Client) RESTConfig() *rest.Config {
	return rest.CopyConfig(c.config)
}
