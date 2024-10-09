package proxy

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func parseAddresses(endpoint *corev1.Endpoints) []string {
	if endpoint == nil {
		return nil
	}
	addresses := make([]string, 0, len(endpoint.Subsets))
	for _, subset := range endpoint.Subsets {
		portNumber := 6443
		for _, port := range subset.Ports {
			if port.Name == "https" {
				portNumber = int(port.Port)
				break
			}
		}

		for _, addr := range subset.Addresses {
			addresses = append(addresses, fmt.Sprintf("%s:%d", addr.IP, portNumber))
		}
	}

	sort.Strings(addresses)
	return addresses
}

func getKubernetesEndpoints(ctx context.Context, kubeconfigFile string) ([]string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubernetes client: %w", err)
	}

	endpoint, err := clientset.CoreV1().Endpoints("default").Get(ctx, "kubernetes", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve endpoints for kubernetes service: %w", err)
	}
	return parseAddresses(endpoint), nil
}
