package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/api"
	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/microcluster/config"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
)

// Config defines configuration for the k8sd app.
type Config struct {
	// Debug increases log message verbosity.
	Debug bool
	// Verbose increases log message verbosity.
	Verbose bool
	// StateDir is the local directory to store the state of the node.
	StateDir string
	// Snap is the snap instance to use.
	Snap snap.Snap
}

// App is the k8sd microcluster instance.
type App struct {
	microCluster *microcluster.MicroCluster
	snap         snap.Snap

	// readyWg is used to denote that the microcluster node is now running
	readyWg sync.WaitGroup

	nodeConfigController *controllers.NodeConfigurationController
}

// New initializes a new microcluster instance from configuration.
func New(ctx context.Context, cfg Config) (*App, error) {
	if cfg.StateDir == "" {
		cfg.StateDir = cfg.Snap.K8sdStateDir()
	}
	cluster, err := microcluster.App(ctx, microcluster.Args{
		Verbose:  cfg.Verbose,
		Debug:    cfg.Debug,
		StateDir: cfg.StateDir,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create microcluster app: %w", err)
	}

	app := &App{
		microCluster: cluster,
		snap:         cfg.Snap,
	}
	app.readyWg.Add(1)

	app.nodeConfigController = controllers.NewNodeConfigurationController(
		cfg.Snap,
		app.readyWg.Wait,
		func() (*k8s.Client, error) {
			return k8s.NewClient(cfg.Snap.KubernetesNodeRESTClientGetter("kube-system"))
		},
	)

	return app, nil
}

// Run starts the microcluster node and waits until it terminates.
// any non-nil customHooks override the default hooks.
func (a *App) Run(customHooks *config.Hooks) error {
	// TODO: consider improving API for overriding hooks.
	hooks := &config.Hooks{
		OnBootstrap: a.onBootstrap,
		PostJoin:    a.onPostJoin,
		PreRemove:   a.onPreRemove,
		OnStart:     a.onStart,
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
		if customHooks.OnStart != nil {
			hooks.OnStart = customHooks.OnStart
		}
	}
	err := a.microCluster.Start(api.New(a).Endpoints(), database.SchemaExtensions, hooks)
	if err != nil {
		return fmt.Errorf("failed to run microcluster: %w", err)
	}
	return nil
}

func (a *App) markNodeReady(ctx context.Context, s *state.State) {
	for {
		if s.Database.IsOpen() {
			a.readyWg.Done()
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}
