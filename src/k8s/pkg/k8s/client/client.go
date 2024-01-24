package client

import (
	"context"
	"fmt"

	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/microcluster"
)

// ClusterOpts contains options for cluster queries.
type ClusterOpts struct {
	// StateDir is the directory that contains the cluster state (for local clients).
	StateDir string
	// Verbose enables info level logging.
	Verbose bool
	// Debug enables trace level logging.
	Debug bool
}

// Client interacts with the k8s REST-API via unix-socket or HTTPS
type Client struct {
	opts ClusterOpts
	m    *microcluster.MicroCluster
	mc   *client.Client
}

// NewClient returns a client to interact with the k8s REST-API
// On a cluster node it will return a client connected to the unix-socket
// elsewhere it returns a HTTPS client that expects the certificates to be located at ClusterOpts.StateDir
func NewClient(ctx context.Context, opts ClusterOpts) (*Client, error) {
	m, err := microcluster.App(ctx, microcluster.Args{
		Debug:    opts.Debug,
		StateDir: opts.StateDir,
		Verbose:  opts.Verbose,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot read cluster info: %w", err)
	}

	microClient, err := m.LocalClient()
	if err != nil {
		return nil, fmt.Errorf("cannot create local client: %w", err)
	}

	return &Client{
		opts: opts,
		m:    m,
		mc:   microClient,
	}, nil
}
