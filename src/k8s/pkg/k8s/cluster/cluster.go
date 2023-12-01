package cluster

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/microcluster"
)

const (
	// DefaultPort is the port under which
	// the REST API is exposed by default.
	DefaultPort = 6400
)

// Client is a wrapper around the MicroCluster client
type Client struct {
	opts ClusterOpts
	app  *microcluster.MicroCluster
}

// ClusterMember holds information about a server in a cluster.
// This is a wrapper around the internal microcluster ClusterMember type.
type ClusterMember struct {
	Name        string
	Address     string
	Role        string
	Fingerprint string
	Status      string
}

// ClusterOpts contains options for cluster queries.
type ClusterOpts struct {
	// StorageDir is the directory that contains the cluster state (for local clients).
	StorageDir string
	// RemoteAddress is the address of the cluster (for remote clients).
	RemoteAddress string
	// Port is the port on which the REST-API is exposed.
	Port string
	// Verbose enables info level logging.
	Verbose bool
	// Debug enables trace level logging.
	Debug bool
}

// NewClient returns a client to interact with the cluster.
// It will return:
//   - a local client, if executing node is part of cluster (valid config.stateDir)
//   - a remote client, if executing node is part of cluster but config.Address is set
//
// TODO(bschimke):
//   - a REST client, if executed from outside of a cluster. A REST client has limited functionality.
//     This requires a mechanism to distribute the server certificates to the client
func NewClient(ctx context.Context, opts ClusterOpts) (*Client, error) {
	err := opts.isValid()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	m, err := microcluster.App(ctx, microcluster.Args{
		Debug:      opts.Debug,
		ListenPort: opts.Port,
		StateDir:   opts.StorageDir,
		Verbose:    opts.Verbose,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot read cluster info: %w", err)
	}

	return &Client{
		opts: opts,
		app:  m,
	}, nil
}

func (c *Client) microClient(ctx context.Context) (*client.Client, error) {
	var microClient *client.Client
	var err error
	if c.opts.RemoteAddress != "" {
		microClient, err = c.app.RemoteClient(c.opts.RemoteAddress)
	} else {
		microClient, err = c.app.LocalClient()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return microClient, nil
}

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

// isValid verifies that a valid IP address is set (for remote)
// and that the state dir exists.
func (c ClusterOpts) isValid() error {
	if c.RemoteAddress != "" {
		if net.ParseIP(c.RemoteAddress) == nil {
			return fmt.Errorf("%s is not a valid IP address", c.RemoteAddress)
		}
		return nil
	}

	if c.StorageDir != "" {
		if _, err := os.Stat(c.StorageDir); os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", c.StorageDir)
		}
		return nil
	}

	return fmt.Errorf("Neither cluster address nor local state dir is set")
}
