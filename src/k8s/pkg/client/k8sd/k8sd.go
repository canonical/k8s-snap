package k8sd

import (
	"fmt"

	"github.com/canonical/microcluster/v2/client"
	"github.com/canonical/microcluster/v2/microcluster"
)

// k8sd implements Client.
type k8sd struct {
	app    *microcluster.MicroCluster
	client *client.Client
}

// New creates a new k8sd client.
// stateDir is the root k8sd state directory, where server.crt and server.key certificates are found.
// address must be left empty to interact with k8sd using the local unix socket from stateDir.
func New(stateDir string) (*k8sd, error) {
	app, err := microcluster.App(microcluster.Args{
		StateDir: stateDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize microcluster app: %w", err)
	}

	client, err := app.LocalClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create local microcluster client: %w", err)
	}

	return &k8sd{
		app:    app,
		client: client,
	}, nil
}

var _ Client = &k8sd{}
