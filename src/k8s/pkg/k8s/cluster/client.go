package cluster

import (
	"context"
	"fmt"
	"net"
	"os"

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

	// TODO: use app.New() instead
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
