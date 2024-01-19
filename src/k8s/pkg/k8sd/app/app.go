package app

import (
	"context"
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/k8sd/api"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/config"
	"github.com/canonical/microcluster/microcluster"
)

// Config defines configuration for the k8sd app.
type Config struct {
	// Debug increases log message verbosity.
	Debug bool
	// Verbose increases log message verbosity.
	Verbose bool
	// ListenPort is the network port to bind for connections.
	ListenPort uint
	// StateDir is the local directory to store the state of the node.
	StateDir string
}

// App is the k8sd microcluster instance.
type App struct {
	MicroCluster *microcluster.MicroCluster
}

// New initializes a new microcluster instance from configuration.
func New(ctx context.Context, cfg Config) (*App, error) {
	snapCtx := snap.ContextWithSnap(ctx, snap.NewSnap(
		os.Getenv("SNAP"),
		os.Getenv("SNAP_DATA"),
		os.Getenv("SNAP_COMMON"),
	))

	cluster, err := microcluster.App(snapCtx, microcluster.Args{
		Verbose:    cfg.Verbose,
		Debug:      cfg.Debug,
		ListenPort: fmt.Sprintf("%d", cfg.ListenPort),
		StateDir:   cfg.StateDir,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create microcluster app: %w", err)
	}

	return &App{
		MicroCluster: cluster,
	}, nil
}

// Run starts the microcluster node and waits until it terminates.
func (a *App) Run(hooks *config.Hooks) error {
	// TODO: define endpoints, schema migrations, hooks
	err := a.MicroCluster.Start(api.Endpoints, database.SchemaExtensions, hooks)
	if err != nil {
		return fmt.Errorf("failed to run microcluster: %w", err)
	}
	return nil
}
