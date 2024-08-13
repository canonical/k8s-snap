package k8sd

import (
	"context"
	"errors"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/lxd/shared/api"
)

func (c *k8sd) NodeStatus(ctx context.Context) (apiv1.NodeStatusResponse, bool, error) {
	response, err := query(ctx, c, "GET", apiv1.NodeStatusRPC, nil, &apiv1.NodeStatusResponse{})
	if err != nil {
		// Error 503 means the node is not initialized yet
		var statusErr api.StatusError
		if errors.As(err, &statusErr) {
			if statusErr.Status() == http.StatusServiceUnavailable {
				return apiv1.NodeStatusResponse{}, false, nil
			}
		}

		return apiv1.NodeStatusResponse{}, false, err
	}

	return response, true, nil
}

func (c *k8sd) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatusResponse, error) {
	var response apiv1.ClusterStatusResponse
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		var err error
		response, err = query(ctx, c, "GET", apiv1.ClusterStatusRPC, nil, &apiv1.ClusterStatusResponse{})
		if err != nil {
			return false, err
		}
		return !waitReady || response.ClusterStatus.Ready, nil
	}); err != nil {
		return apiv1.ClusterStatusResponse{}, err
	}
	return response, nil
}
