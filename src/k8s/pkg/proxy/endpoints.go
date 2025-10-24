package proxy

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubernetesEndpoints(ctx context.Context, kubeconfigFile string, server string) ([]string, error) {
	config, err := clientcmd.BuildConfigFromFlags(fmt.Sprintf("https://%s", server), kubeconfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read load kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubernetes client: %w", err)
	}

	endpoints, err := clientset.CoreV1().Endpoints("default").Get(ctx, "kubernetes", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve endpoints for kubernetes service: %w", err)
	}
	if endpoints == nil {
		return nil, nil
	}

	return utils.ParseEndpoints(endpoints), nil
}
