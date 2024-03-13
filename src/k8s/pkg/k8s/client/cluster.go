package client

import (
	"context"
	"fmt"
	"os"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/lxd/shared/api"
)

// IsBootstrapped checks if the cluster is already up and initialized.
func (c *k8sdClient) IsBootstrapped(ctx context.Context) bool {
	_, err := c.m.Status()
	return err == nil
}

// Bootstrap bootstraps the k8s cluster
func (c *k8sdClient) Bootstrap(ctx context.Context, hostname string, address string, bootstrapConfig apiv1.BootstrapConfig) (apiv1.NodeStatus, error) {
	timeout := 30 * time.Second
	if deadline, set := ctx.Deadline(); set {
		timeout = time.Until(deadline)
	}

	if err := c.m.Ready(int(timeout / time.Second)); err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("k8sd API is not ready: %w", err)
	}
	config, err := bootstrapConfig.ToMap()
	if err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("failed to convert bootstrap config to map: %w", err)
	}
	if err := c.m.NewCluster(hostname, address, config, timeout); err != nil {
		// TODO(neoaggelos): only return error that bootstrap failed
		fmt.Fprintln(os.Stderr, "Failed with error:", err)
		c.CleanupNode(ctx, hostname)
		return apiv1.NodeStatus{}, fmt.Errorf("failed to bootstrap new cluster: %w", err)
	}

	// TODO(neoaggelos): retrieve hostname and address from the cluster, do not guess
	return apiv1.NodeStatus{
		Name:    hostname,
		Address: address,
	}, nil
}

// ClusterStatus returns the current status of the cluster.
func (c *k8sdClient) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
	var response apiv1.GetClusterStatusResponse

	if !waitReady && !c.IsKubernetesAPIServerReady(ctx) {
		return apiv1.ClusterStatus{}, fmt.Errorf("there are no active kube-apiserver endpoints, cluster status is unavailable")
	}

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

// IsKubernetesAPIServerReady checks if kube-apiserver is reachable.
func (c *k8sdClient) IsKubernetesAPIServerReady(ctx context.Context) bool {
	kc, err := k8s.NewClient(c.snap)
	if err != nil {
		return false
	}
	_, err = kc.GetKubeAPIServerEndpoints(ctx)
	if err != nil {
		return false
	}
	return err == nil
}
