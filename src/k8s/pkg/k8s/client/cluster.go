package client

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/lxd/util"
	lxdApi "github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/microcluster"
)

// Bootstrap Cluster (to be removed)
// TODO: This does not use an REST endpoint because it will eventually move into the k8s init command anyway.
func (c *Client) Bootstrap(ctx context.Context) (api.ClusterMember, error) {
	// Get system hostname.
	hostname, err := os.Hostname()
	if err != nil {
		return api.ClusterMember{}, fmt.Errorf("failed to retrieve system hostname: %w", err)
	}

	port, err := strconv.Atoi(c.opts.Port)
	if err != nil {
		return api.ClusterMember{}, fmt.Errorf("failed to parse Port: %w", err)
	}
	// Get system address.
	address := util.CanonicalNetworkAddress(
		util.NetworkInterfaceAddress(), port,
	)

	member := api.ClusterMember{
		Name:    hostname,
		Address: address,
	}
	m, err := microcluster.App(ctx, microcluster.Args{StateDir: c.opts.StorageDir, Verbose: false, Debug: false})
	if err != nil {
		return api.ClusterMember{}, fmt.Errorf("failed to configure MicroCluster: %w", err)
	}
	err = m.NewCluster(hostname, address, time.Second*30)
	return member, err
}

// ClusterStatus returns the current status of the cluster.
func (c *Client) ClusterStatus(ctx context.Context) (api.ClusterStatus, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response api.GetClusterStatusResponse
	err := c.mc.Query(queryCtx, "GET", lxdApi.NewURL().Path("k8sd", "cluster"), nil, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return api.ClusterStatus{}, fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return response.ClusterStatus, nil
}
