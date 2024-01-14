package utils

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeClient wraps a Kubernetes clientset to provide simplified interactions.
type KubeClient struct {
	clientSet *kubernetes.Clientset
}

// NewKubeClient creates a new Kubernetes client from a given kubeconfig file path.
func NewKubeClient(kubeconfigPath string) (*KubeClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &KubeClient{clientSet: clientset}, nil
}

// GetService retrieves a Kubernetes service by name and namespace.
// An empty namespace will default to "default".
func (kc *KubeClient) GetService(ctx context.Context, name, namespace string) (*v1.Service, error) {
	if namespace == "" {
		namespace = "default"
	}

	svc, err := kc.clientSet.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service '%s' in namespace '%s': %w", name, namespace, err)
	}

	return svc, nil
}
