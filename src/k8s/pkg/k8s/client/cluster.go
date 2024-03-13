package client

import (
	"context"
	"fmt"
	"os"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/lxd/shared/api"
)

// IsBootstrapped checks if the cluster is already up and initialized.
func (c *k8sdClient) IsBootstrapped(ctx context.Context) bool {
	_, err := c.m.Status()
	return err == nil
}

// Bootstrap bootstraps the k8s cluster
func (c *k8sdClient) Bootstrap(ctx context.Context, bootstrapConfig apiv1.BootstrapConfig) (apiv1.NodeStatus, error) {
	// Get system hostname.
	rawHostname, err := os.Hostname()
	if err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("failed to retrieve system hostname: %w", err)
	}
	// TODO: this should be done on the server side, but we cannot currently hijack the microcluster bootstrap endpoint.
	hostname, err := utils.CleanHostname(rawHostname)
	if err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("invalid hostname %q: %w", rawHostname, err)
	}

	// Get system addrPort.
	addrPort := util.CanonicalNetworkAddress(util.NetworkInterfaceAddress(), config.DefaultPort)

	timeout := 30 * time.Second
	if deadline, set := ctx.Deadline(); set {
		timeout = time.Until(deadline)
	}

	if err := c.m.Ready(int(timeout / time.Second)); err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("cluster did not come up in time: %w", err)
	}
	config, err := bootstrapConfig.ToMap()
	if err != nil {
		return apiv1.NodeStatus{}, fmt.Errorf("failed to convert bootstrap config to map: %w", err)
	}
	if err := c.m.NewCluster(hostname, addrPort, config, timeout); err != nil {
		// TODO(neoaggelos): only return error that bootstrap failed
		fmt.Fprintln(os.Stderr, "Failed with error:", err)
		c.CleanupNode(ctx, hostname)
		return apiv1.NodeStatus{}, fmt.Errorf("failed to bootstrap new cluster: %w", err)
	}

	// TODO(neoaggelos): retrieve hostname and address from the cluster, do not guess
	return apiv1.NodeStatus{
		Name:    hostname,
		Address: util.NetworkInterfaceAddress(),
	}, nil
}

// ClusterStatus returns the current status of the cluster.
func (c *k8sdClient) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
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
func (c *k8sdClient) KubeConfig(ctx context.Context, server string) (string, error) {
	request := apiv1.GetKubeConfigRequest{Server: server}
	response := apiv1.GetKubeConfigResponse{}

	err := c.Query(ctx, "GET", api.NewURL().Path("k8sd", "kubeconfig"), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return "", fmt.Errorf("failed to query endpoint GET /k8sd/kubeconfig on %q: %w", clientURL.String(), err)
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
