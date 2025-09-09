package proxy

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubernetesEndpoints(ctx context.Context, kubeconfigFile string) ([]string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubernetes client: %w", err)
	}

	endpointSlices, err := clientset.DiscoveryV1().EndpointSlices("default").List(ctx, metav1.ListOptions{
		LabelSelector: "kubernetes.io/service-name=kubernetes",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve endpoints for kubernetes service: %w", err)
	}
	if endpointSlices == nil {
		return nil, nil
	}

	return utils.ParseEndpoints(endpointSlices), nil
}
