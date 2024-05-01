package kubernetes

import (
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// Client is a wrapper around the kubernetes.Interface.
type Client struct {
	kubernetes.Interface
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
	return &Client{clientset}, nil
}
