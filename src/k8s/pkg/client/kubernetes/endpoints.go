package kubernetes

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/utils"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// GetKubeAPIServerEndpoints retrieves the known kube-apiserver endpoints of the cluster.
// GetKubeAPIServerEndpoints returns an error if the list of endpoints is empty.
func (c *Client) GetKubeAPIServerEndpoints(ctx context.Context) ([]string, error) {
	var endpointSlices *discoveryv1.EndpointSliceList
	var err error
	err = retry.OnError(retry.DefaultBackoff, func(err error) bool { return true }, func() error {
		endpointSlices, err = c.DiscoveryV1().EndpointSlices("default").List(ctx, metav1.ListOptions{
			LabelSelector: "kubernetes.io/service-name=kubernetes",
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints for kubernetes service: %w", err)
	}
	if endpointSlices == nil {
		return nil, fmt.Errorf("endpoints for kubernetes service not found")
	}

	addresses := utils.ParseEndpoints(endpointSlices)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("empty list of endpoints for the kubernetes service")
	}

	return addresses, nil
}
