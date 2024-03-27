package client

import (
	"context"
	"fmt"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/lxd/shared/api"
)

// IsBootstrapped checks if the cluster is already up and initialized.
func (c *k8sdClient) IsBootstrapped(ctx context.Context) bool {
	_, err := c.m.Status()
	return err == nil
}

// Bootstrap bootstraps the k8s cluster
func (c *k8sdClient) Bootstrap(ctx context.Context, request apiv1.PostClusterBootstrapRequest) (apiv1.NodeStatus, error) {
	timeout := utils.TimeoutFromCtx(ctx)

	if err := c.m.Ready(int(timeout / time.Second)); err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("k8sd API is not ready: %w", err)
	}
	response := apiv1.NodeStatus{}

	if err := c.mc.Query(ctx, "POST", api.NewURL().Path("k8sd", "cluster"), request, &response); err != nil {
		c.CleanupNode(ctx, request.Name)
		return response, fmt.Errorf("failed to bootstrap new cluster using POST /k8sd/cluster: %w", err)
	}

	return response, nil
}

// ClusterStatus returns the current status of the cluster.
func (c *k8sdClient) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
	var response apiv1.GetClusterStatusResponse
	err := control.WaitUntilReady(ctx, func() (bool, error) {
		if err := c.mc.Query(ctx, "GET", api.NewURL().Path("k8sd", "cluster"), nil, &response); err != nil {
			return false, fmt.Errorf("failed to GET /k8sd/cluster: %w", err)
		}
		return !waitReady || response.ClusterStatus.Ready, nil
	})
	return response.ClusterStatus, err
}

// KubeConfig returns admin kubeconfig to connect to the cluster.
func (c *k8sdClient) KubeConfig(ctx context.Context, request apiv1.GetKubeConfigRequest) (string, error) {
	response := apiv1.GetKubeConfigResponse{}
	if err := c.mc.Query(ctx, "GET", api.NewURL().Path("k8sd", "kubeconfig"), request, &response); err != nil {
		return "", fmt.Errorf("failed to GET /k8sd/kubeconfig: %w", err)
	}
	return response.KubeConfig, nil
}
