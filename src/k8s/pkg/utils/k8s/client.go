package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// k8sClient is a wrapper around the k8s clientset
type k8sClient struct {
	kubernetes.Interface
}

// NewClient creates a client to the k8s cluster.
//
// TODO:
// There is no way for the user to overwrite this kubeconfig.
// We might need to add this functionality similar to `k8s kubectl`.
// However, simply querying the KUBECONFIG env will not work for remote clients.
func NewClient() (*k8sClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/kubernetes/admin.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clientset: %w", err)
	}

	return &k8sClient{
		clientset,
	}, nil
}
