package app

import (
	"context"
	"fmt"

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
	snapCtx := snap.ContextWithSnap(ctx, snap.NewDefaultSnap())

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
// any non-nil customHooks override the default hooks.
func (a *App) Run(customHooks *config.Hooks) error {
	// TODO: consider improving API for overriding hooks.
	hooks := &config.Hooks{
		OnBootstrap: onBootstrap,
		PostJoin:    onPostJoin,
		PreRemove:   onPreRemove,
		OnNewMember: onNewMember,
	}
	if customHooks != nil {
		if customHooks.OnBootstrap != nil {
			hooks.OnBootstrap = customHooks.OnBootstrap
		}
		if customHooks.PostJoin != nil {
			hooks.PostJoin = customHooks.PostJoin
		}
		if customHooks.PreRemove != nil {
			hooks.PreRemove = customHooks.PreRemove
		}
	}
	err := a.MicroCluster.Start(api.Endpoints(a.MicroCluster), database.SchemaExtensions, hooks)
	if err != nil {
		return fmt.Errorf("failed to run microcluster: %w", err)
	}
	return nil
}
