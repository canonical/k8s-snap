package client

import (
	"context"
	"fmt"
	"os"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/lxd/shared/api"
)

// IsBootstrapped checks if the cluster is already up and initialized.
func (c *Client) IsBootstrapped(ctx context.Context) bool {
	_, err := c.m.Status()
	return err == nil
}

// Bootstrap bootstraps the k8s cluster
func (c *Client) Bootstrap(ctx context.Context, bootstrapConfig apiv1.BootstrapConfig) (apiv1.ClusterMember, error) {
	// Get system hostname.
	hostname, err := os.Hostname()
	if err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("failed to retrieve system hostname: %w", err)
	}

	// Get system addrPort.
	addrPort := util.CanonicalNetworkAddress(util.NetworkInterfaceAddress(), config.DefaultPort)

	timeToWait := 30
	// If a context timeout is set, use this instead.
	deadline, set := ctx.Deadline()
	if set {
		timeToWait = int(deadline.Sub(time.Now()).Seconds())
	}

	if err := c.m.Ready(timeToWait); err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("cluster did not come up in time: %w", err)
	}
	config, err := bootstrapConfig.ToMap()
	if err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("failed to convert bootstrap config to map: %w", err)
	}
	if err := c.m.NewCluster(hostname, addrPort, config, time.Second*30); err != nil {
		// TODO(neoaggelos): print message that bootstrap failed, and that we are cleaning up
		fmt.Fprintln(os.Stderr, "Failed with error:", err)
		c.CleanupNode(ctx, c.opts.Snap, hostname)
		return apiv1.ClusterMember{}, fmt.Errorf("failed to bootstrap new cluster: %w", err)
	}

	// TODO(neoaggelos): retrieve hostname and address from the cluster, do not guess
	return apiv1.ClusterMember{
		Name:    hostname,
		Address: util.NetworkInterfaceAddress(),
	}, nil
}

// ClusterStatus returns the current status of the cluster.
func (c *Client) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
	var response apiv1.GetClusterStatusResponse
	err := control.WaitUntilReady(ctx, func() (bool, error) {
		err := c.Query(ctx, "GET", api.NewURL().Path("k8sd", "cluster"), nil, &response)
		if err != nil {
			return false, err
		}
		return !waitReady || response.ClusterStatus.Ready, nil
	})
	return response.ClusterStatus, err
}

// KubeConfig returns admin kubeconfig to connect to the cluster.
func (c *Client) KubeConfig(ctx context.Context) (string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response apiv1.GetKubeConfigResponse
	err := c.Query(queryCtx, "GET", api.NewURL().Path("k8sd", "kubeconfig"), nil, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return "", fmt.Errorf("failed to query endpoint GET /k8sd/kubeconfig on %q: %w", clientURL.String(), err)
	}
	return response.KubeConfig, nil
}
