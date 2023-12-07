package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/microcluster"
)

const (
	// DefaultPort is the port under which
	// the REST API is exposed by default.
	DefaultPort = 6400
)

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

// Client interacts with the k8s REST-API via unix-socket or HTTPS
type Client struct {
	opts ClusterOpts
	mc   *client.Client
}

// NewClient returns a client to interact with the k8s REST-API
// On a cluster node it will return a client connected to the unix-socket
// elsewhere it returns a HTTPS client that expects the certificates to be located at ClusterOpts.StorageDir
func NewClient(ctx context.Context, opts ClusterOpts) (*Client, error) {
	m, err := microcluster.App(ctx, microcluster.Args{
		Debug:      opts.Debug,
		ListenPort: opts.Port,
		StateDir:   opts.StorageDir,
		Verbose:    opts.Verbose,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot read cluster info: %w", err)
	}

	var microClient *client.Client
	if opts.RemoteAddress == "" {
		microClient, err = m.LocalClient()
		if err != nil {
			return nil, fmt.Errorf("cannot create local client: %w", err)
		}
	} else {
		// TODO: Implement the remote client. This requires the cluster certs to be available at `opts.StorageDir`
		return nil, errors.New("remote clients are not yet supported. The CLI needs to run on a cluster node.")
	}

	return &Client{
		opts: opts,
		mc:   microClient,
	}, nil
}
