package k8sd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api-v1/api/v1"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sd) NodeStatus(ctx context.Context) (apiv1.NodeStatus, bool, error) {
	var response apiv1.NodeStatusResponse
	if err := c.client.Query(ctx, "GET", apiv1.K8sdAPIVersion, api.NewURL().Path("k8sd", "node"), nil, &response); err != nil {

		// Error 503 means the node is not initialized yet
		var statusErr api.StatusError
		if errors.As(err, &statusErr) {
			if statusErr.Status() == http.StatusServiceUnavailable {
				return apiv1.NodeStatus{}, false, nil
			}
		}

		return apiv1.NodeStatus{}, false, fmt.Errorf("failed to GET /k8sd/node: %w", err)
	}
	return response.NodeStatus, true, nil
}

func (c *k8sd) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
	var response apiv1.ClusterStatusResponse
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		if err := c.client.Query(ctx, "GET", apiv1.K8sdAPIVersion, api.NewURL().Path("k8sd", "cluster"), nil, &response); err != nil {
			return false, fmt.Errorf("failed to GET /k8sd/cluster: %w", err)
		}
		return !waitReady || response.ClusterStatus.Ready, nil
	}); err != nil {
		return apiv1.ClusterStatus{}, err
	}
	return response.ClusterStatus, nil
}
