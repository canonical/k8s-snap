package client

import (
	"context"
	"fmt"
	"time"

	api "github.com/canonical/k8s/api/v1"
	lxdApi "github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/microcluster"
)

// JoinNode calls "POST 1.0/k8sd/cluster/<node>"
func (c *Client) JoinNode(ctx context.Context, name string, address string, token string) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	// TODO: This is super ugly but we first need to "initialize" the database with this join command before we can
	// access the REST-API.
	// This 'workaround' joins the microcluster to the bootstrapped microcluster and then calls our own '/clustering' endpoint -ugh
	// This will break if we do not have a k8sd instance running on this node.
	// Some notes:
	// (1) Find a way to access the REST-API /clustering endpoint before "init" the DB (if we try to access before it fails with "daemon not initialized")
	// (2) we cannot use the hooks that microcluster provides (e.g. onJoinMember) as we require additional req/resp data that are not covered by the /cluster endpoint
	// (3) I tried to bootstrap nodes independently (having two independent clusters) and then join the one to the other by calling our own `/clustering`:
	// 		(a) simply trying this causes 'node' does already exist errors
	// 		(b) Tried to remove all nodes from the joining node but this fails as a cluster cannot have zero members.
	m, err := microcluster.App(ctx, microcluster.Args{StateDir: c.opts.StorageDir, Verbose: false, Debug: false})
	if err != nil {
		return fmt.Errorf("failed to configure MicroCluster: %w", err)
	}

	err = m.JoinCluster(name, address, token, time.Second*10)
	if err != nil {
		return fmt.Errorf("failed to join node %s to cluster: %w", name, err)
	}

	request := api.AddNodeRequest{
		Address: address,
		Token:   token,
	}
	var response api.AddNodeResponse
	err = c.mc.Query(queryCtx, "POST", lxdApi.NewURL().Path("k8sd", "cluster", name), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return nil
}

// RemoveNode calls "DELETE 1.0/k8sd/cluster/<node>"
func (c *Client) RemoveNode(ctx context.Context, name string, force bool) error {
	queryCtx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	request := api.RemoveNodeRequest{
		Force: force,
	}
	var response api.RemoveNodeResponse
	err := c.mc.Query(queryCtx, "DELETE", lxdApi.NewURL().Path("k8sd", "cluster", name), request, &response)
	if err != nil {
		clientURL := c.mc.URL()
		return fmt.Errorf("failed to query endpoint on %q: %w", clientURL.String(), err)
	}
	return nil
}
