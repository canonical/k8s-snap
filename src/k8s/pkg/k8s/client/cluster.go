package client

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/microcluster"
)

// Bootstrap Cluster (to be removed)
// TODO: This does not use an REST endpoint because it will eventually move into the k8s init command anyway.
func (c *Client) Bootstrap(ctx context.Context) (apiv1.ClusterMember, error) {
	// Get system hostname.
	hostname, err := os.Hostname()
	if err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("failed to retrieve system hostname: %w", err)
	}

	port, err := strconv.Atoi(c.opts.Port)
	if err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("failed to parse Port: %w", err)
	}
	// Get system address.
	address := util.CanonicalNetworkAddress(
		util.NetworkInterfaceAddress(), port,
	)

	member := apiv1.ClusterMember{
		Name:    hostname,
		Address: address,
	}
	m, err := microcluster.App(ctx, microcluster.Args{StateDir: c.opts.StorageDir, Verbose: false, Debug: false})
	if err != nil {
		return apiv1.ClusterMember{}, fmt.Errorf("failed to configure MicroCluster: %w", err)
	}
	err = m.NewCluster(hostname, address, time.Second*30)

	// Make init cluster call to REST endpoint
	// TODO: Right now this only takes care of storing k8s-dqlite certificates in k8sd
	//       Eventually we need to move all the k8s init code to the REST api
	//       and drop k8s bootstrap-cluster
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var response apiv1.GetClusterStatusResponse
	err = c.mc.Query(queryCtx, "POST", api.NewURL().Path("k8sd", "cluster"), nil, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return apiv1.ClusterMember{}, fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}

	return member, err
}

// ClusterStatus returns the current status of the cluster.
func (c *Client) ClusterStatus(ctx context.Context) (apiv1.ClusterStatus, error) {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var response apiv1.GetClusterStatusResponse
	err := c.mc.Query(queryCtx, "GET", api.NewURL().Path("k8sd", "cluster"), nil, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return apiv1.ClusterStatus{}, fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return response.ClusterStatus, nil
}
