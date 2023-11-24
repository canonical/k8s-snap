package cluster

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/canonical/lxd/lxd/util"
	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/microcluster"
)

// Boostrap sets up new cluster
func Bootstrap(ctx context.Context, config ClusterOpts) error {
	m, err := getMicroClusterApp(ctx, config)
	if err != nil {
		return fmt.Errorf("unable to configure cluster: %w", err)
	}

	// Get system hostname.
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to retrieve system hostname: %w", err)
	}

	// Get system address.
	address := util.NetworkInterfaceAddress()
	address = util.CanonicalNetworkAddress(address, 6443)

	return m.NewCluster(hostname, address, time.Second*30)
}

func GetMembers(ctx context.Context, config ClusterOpts) ([]ClusterMember, error) {
	client, err := getClient(ctx, config)
	if err != nil {
		return nil, err
	}

	clusterMembers, err := client.GetClusterMembers(context.Background())
	if err != nil {
		return nil, err
	}

	members := make([]ClusterMember, len(clusterMembers))
	for i, clusterMember := range clusterMembers {
		fingerprint, err := shared.CertFingerprintStr(clusterMember.Certificate.String())
		if err != nil {
			continue
		}

		members[i] = ClusterMember{
			clusterMember.Name,
			clusterMember.Address.String(),
			clusterMember.Role,
			fingerprint,
			string(clusterMember.Status),
		}
	}

	return members, nil
}

// Returns client to interact with the cluster.
// It will return:
//   - a local client, if executing node is part of cluster (valid config.stateDir)
//   - a remote client, if executing node is part of cluster but config.Address is set
//
// TODO(bschimke):
//   - a REST client, if executed from outside of a cluster. A REST client has limited functionality.
//     This requires a mechanism to distribute the server certificates to the client
func getClient(ctx context.Context, config ClusterOpts) (*client.Client, error) {
	err := config.isValid()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	m, err := getMicroClusterApp(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("cannot create client: %w", err)
	}

	if config.Address != "" {
		return m.RemoteClient(config.Address)
	}

	return m.LocalClient()
}

func getMicroClusterApp(ctx context.Context, config ClusterOpts) (*microcluster.MicroCluster, error) {
	m, err := microcluster.App(ctx, microcluster.Args{
		StateDir: config.StateDir,
		Verbose:  config.Verbose,
		Debug:    config.Debug,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot read cluster info: %w", err)
	}
	return m, nil
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
	// Directory containing cluster information (for local clients)
	// TODO(bschimke): We can set the snaps default directory here once we snapped the CLI
	StateDir string
	// Cluster address (for remote clients)
	Address string
	// Enable info level logging
	Verbose bool
	// Enable trace level logging
	Debug bool
}

// isValid verifies that a valid IP address is set (for remote)
// and that the state dir exists.
func (c ClusterOpts) isValid() error {
	if c.Address != "" {
		if net.ParseIP(c.Address) == nil {
			return fmt.Errorf("%s is not a valid IP address", c.Address)
		}
		return nil
	}

	if c.StateDir != "" {
		if _, err := os.Stat(c.StateDir); os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", c.StateDir)
		}
		_, err := net.Dial("unix", filepath.Join(c.StateDir, "control.socket"))
		if err != nil {
			return fmt.Errorf("cannot connect to local cluster - is it running?")
		}
		return nil
	}

	return fmt.Errorf("Neither cluster address nor local state dir is set")
}
