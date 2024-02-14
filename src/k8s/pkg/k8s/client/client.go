package client

import (
	"context"
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/microcluster"
)

// ClusterOpts contains options for cluster queries.
type ClusterOpts struct {
	// Verbose enables info level logging.
	Verbose bool
	// Debug enables trace level logging.
	Debug bool
	// Snap is the snap instance.
	Snap snap.Snap
	// StateDir is the directory that contains the cluster state (for local clients).
	StateDir string
}

// k8sdClient interacts with the k8s REST-API via unix-socket or HTTPS
type k8sdClient struct {
	opts ClusterOpts
	m    *microcluster.MicroCluster
	mc   *client.Client
	snap snap.Snap
}

// NewClient returns a client to interact with the k8s REST-API
// On a cluster node it will return a client connected to the unix-socket
// elsewhere it returns a HTTPS client that expects the certificates to be located at ClusterOpts.StateDir
func NewClient(ctx context.Context, opts ClusterOpts) (*k8sdClient, error) {
	// TODO: pass snap through opts instead, do not create here.
	if opts.Snap == nil {
		opts.Snap = snap.NewSnap(os.Getenv("SNAP"), os.Getenv("SNAP_COMMON"))
	}
	if opts.StateDir == "" {
		opts.StateDir = opts.Snap.K8sdStateDir()
	}
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

	return &k8sdClient{
		opts: opts,
		m:    m,
		mc:   microClient,
	}, nil
}

func (c *k8sdClient) Query(ctx context.Context, method string, path *api.URL, in any, out any) error {
	if err := c.mc.Query(ctx, method, path, in, out); err != nil {
		return err
	}
	return nil
}
