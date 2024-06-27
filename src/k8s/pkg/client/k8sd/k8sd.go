package k8sd

import (
	"fmt"

	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/microcluster"
)

// k8sd implements Client.
type k8sd struct {
	app    *microcluster.MicroCluster
	client *client.Client
}

// New creates a new k8sd client.
// stateDir is the root k8sd state directory, where server.crt and server.key certificates are found.
// address must be left empty to interact with k8sd using the local unix socket from stateDir.
func New(stateDir string, address string) (*k8sd, error) {
	app, err := microcluster.App(microcluster.Args{
		StateDir: stateDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize microcluster app: %w", err)
	}

	var client *client.Client
	if address == "" {
		if client, err = app.LocalClient(); err != nil {
			return nil, fmt.Errorf("failed to create local microcluster client: %w", err)
		}
	} else {
		if client, err = app.RemoteClient(address); err != nil {
			return nil, fmt.Errorf("failed to create remote microcluster client to %q: %w", address, err)
		}
	}

	return &k8sd{
		app:    app,
		client: client,
	}, nil
}

var _ Client = &k8sd{}
