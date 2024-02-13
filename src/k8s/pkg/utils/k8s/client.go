package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
	"k8s.io/client-go/kubernetes"
)

// Client is a wrapper around the kubernetes.Interface.
type Client struct {
	kubernetes.Interface
}

func NewClient(snap snap.Snap) (*Client, error) {
	config, err := snap.KubernetesRESTClientGetter().ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build Kubernetes REST config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}
	return &Client{clientset}, nil
}
