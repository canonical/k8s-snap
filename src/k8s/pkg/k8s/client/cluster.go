package client

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/lxd/shared/api"
)

// Init bootstraps the k8s cluster
func (c *Client) Init(ctx context.Context) (apiv1.ClusterMember, error) {
	// Get system hostname.
	hostname, err := os.Hostname()
	if err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("failed to retrieve system hostname: %w", err)
	}

	port, err := strconv.Atoi(c.opts.Port)
	if err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("failed to parse Port: %w", err)
	}
	// Get system addrPort.
	addrPort := util.CanonicalNetworkAddress(
		util.NetworkInterfaceAddress(), port,
	)

	// This should be done behind the REST API.
	// However, the K8sd daemon needs to be initialized before
	// the REST API can be used.
	// TODO: Find a way to do the bootstrapping/joining of k8sd behind
	//       the REST API.
	err = c.m.NewCluster(hostname, addrPort, time.Second*30)
	if err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("failed to bootstrap new cluster: %w", err)
	}

	err = c.m.Ready(30)
	if err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("cluster did not come up in time: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response apiv1.GetClusterStatusResponse
	err = c.mc.Query(queryCtx, "POST", api.NewURL().Path("k8sd", "cluster"), nil, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return apiv1.ClusterMember{}, fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}

	return apiv1.ClusterMember{
		Name:    hostname,
		Address: util.NetworkInterfaceAddress(),
	}, err
}

// ClusterStatus returns the current status of the cluster.
func (c *Client) ClusterStatus(ctx context.Context, waitReady bool) (apiv1.ClusterStatus, error) {
	var response apiv1.GetClusterStatusResponse
	err := utils.WaitUntilReady(ctx, func() (bool, error) {
		err := c.mc.Query(ctx, "GET", api.NewURL().Path("k8sd", "cluster"), nil, &response)
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
	err := c.mc.Query(queryCtx, "GET", api.NewURL().Path("k8sd", "kubeconfig"), nil, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return "", fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return response.KubeConfig, nil
}
