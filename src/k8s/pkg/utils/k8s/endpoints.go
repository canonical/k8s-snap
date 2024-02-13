package k8s

import (
	"context"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetKubeAPIServerEndpoints retrieves the known kube-apiserver endpoints of the cluster.
// GetKubeAPIServerEndpoints returns an error if the list of endpoints is empty.
func (c *Client) GetKubeAPIServerEndpoints(ctx context.Context) ([]string, error) {
	endpoint, err := c.CoreV1().Endpoints("default").Get(ctx, "kubernetes", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints for kubernetes service: %w", err)
	}
	if endpoint == nil {
		return nil, fmt.Errorf("endpoints for kubernetes service not found")
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
			if addr.IP != "" {
				addresses = append(addresses, fmt.Sprintf("%s:%d", addr.IP, portNumber))
			}
		}
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("empty list of endpoints for the kubernetes service")
	}

	sort.Strings(addresses)
	return addresses, nil
}
