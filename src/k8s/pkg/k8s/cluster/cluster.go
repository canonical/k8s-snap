package cluster

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/lxd/shared"
)

// Bootstrap sets up new cluster and returns the information about the daemon.
func (c *Client) Bootstrap(ctx context.Context) (ClusterMember, error) {
	// Get system hostname.
	hostname, err := os.Hostname()
	if err != nil {
		return ClusterMember{}, fmt.Errorf("failed to retrieve system hostname: %w", err)
	}

	port, err := strconv.Atoi(c.opts.Port)
	if err != nil {
		return ClusterMember{}, fmt.Errorf("failed to parse Port: %w", err)
	}
	// Get system address.
	address := util.CanonicalNetworkAddress(
		util.NetworkInterfaceAddress(), port,
	)

	member := ClusterMember{
		Name:    hostname,
		Address: address,
	}
	err = c.app.NewCluster(hostname, address, time.Second*30)
	return member, err
}

// GetMembers returns information about all members of the cluster.
func (c *Client) GetMembers(ctx context.Context) ([]ClusterMember, error) {
	microClient, err := c.microClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	clusterMembers, err := microClient.GetClusterMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster members: %w", err)
	}

	members := make([]ClusterMember, len(clusterMembers))
	for i, clusterMember := range clusterMembers {
		fingerprint, err := shared.CertFingerprintStr(clusterMember.Certificate.String())
		if err != nil {
			continue
		}

		members[i] = ClusterMember{
			Name:        clusterMember.Name,
			Address:     clusterMember.Address.String(),
			Role:        clusterMember.Role,
			Fingerprint: fingerprint,
			Status:      string(clusterMember.Status),
		}
	}

	return members, nil
}

// GetToken returns a token for a node to use to join the cluster.
func (c *Client) GetToken(ctx context.Context, name string) (string, error) {
	microClient, err := c.microClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get client: %w", err)
	}
	return microClient.RequestToken(ctx, name)
}

// JoinCluster joins a node to an existing cluster (token is supplied by existing cluster member)
func (c *Client) JoinCluster(ctx context.Context, name string, address string, token string) error {
	return c.app.JoinCluster(name, address, token, time.Second*30)
}

// RemoveNode removes a node by name from the cluster
func (c *Client) RemoveNode(ctx context.Context, name string, force bool) error {
	microClient, err := c.microClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}
	err = microClient.DeleteClusterMember(ctx, name, force)
	if err != nil {
		return fmt.Errorf("failed to delete cluster member %s: %w", name, err)
	}
	return nil
}
