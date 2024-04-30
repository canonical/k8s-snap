package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/api"
	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/snap"
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
	// PprofAddress is the address to listen for pprof debug endpoints. Empty to disable.
	PprofAddress string
}

// App is the k8sd microcluster instance.
type App struct {
	microCluster *microcluster.MicroCluster
	snap         snap.Snap

	// profilingAddress
	profilingAddress string

	// readyWg is used to denote that the microcluster node is now running
	readyWg sync.WaitGroup

	nodeConfigController         *controllers.NodeConfigurationController
	controlPlaneConfigController *controllers.ControlPlaneConfigurationController

	// updateNodeConfigController
	triggerUpdateNodeConfigControllerCh chan struct{}
	updateNodeConfigController          *controllers.UpdateNodeConfigurationController

	// featureController
	triggerFeatureControllerNetworkCh       chan struct{}
	triggerFeatureControllerGatewayCh       chan struct{}
	triggerFeatureControllerIngressCh       chan struct{}
	triggerFeatureControllerLoadBalancerCh  chan struct{}
	triggerFeatureControllerLocalStorageCh  chan struct{}
	triggerFeatureControllerMetricsServerCh chan struct{}
	triggerFeatureControllerDNSCh           chan struct{}
	featureController                       *controllers.FeatureController
}

// New initializes a new microcluster instance from configuration.
func New(cfg Config) (*App, error) {
	if cfg.StateDir == "" {
		cfg.StateDir = cfg.Snap.K8sdStateDir()
	}
	cluster, err := microcluster.App(microcluster.Args{
		Verbose:  cfg.Verbose,
		Debug:    cfg.Debug,
		StateDir: cfg.StateDir,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create microcluster app: %w", err)
	}

	app := &App{
		microCluster:     cluster,
		snap:             cfg.Snap,
		profilingAddress: cfg.PprofAddress,
	}
	app.readyWg.Add(1)

	app.nodeConfigController = controllers.NewNodeConfigurationController(
		cfg.Snap,
		app.readyWg.Wait,
	)

	app.controlPlaneConfigController = controllers.NewControlPlaneConfigurationController(
		cfg.Snap,
		app.readyWg.Wait,
		time.NewTicker(10*time.Second).C,
	)

	app.triggerUpdateNodeConfigControllerCh = make(chan struct{}, 1)
	app.updateNodeConfigController = controllers.NewUpdateNodeConfigurationController(
		cfg.Snap,
		app.readyWg.Wait,
		app.triggerUpdateNodeConfigControllerCh,
	)

	app.triggerFeatureControllerNetworkCh = make(chan struct{}, 1)
	app.triggerFeatureControllerGatewayCh = make(chan struct{}, 1)
	app.triggerFeatureControllerIngressCh = make(chan struct{}, 1)
	app.triggerFeatureControllerLoadBalancerCh = make(chan struct{}, 1)
	app.triggerFeatureControllerLocalStorageCh = make(chan struct{}, 1)
	app.triggerFeatureControllerMetricsServerCh = make(chan struct{}, 1)
	app.triggerFeatureControllerDNSCh = make(chan struct{}, 1)
	app.featureController = controllers.NewFeatureController(controllers.FeatureControllerOpts{
		Snap:                   cfg.Snap,
		WaitReady:              app.readyWg.Wait,
		TriggerNetworkCh:       app.triggerFeatureControllerNetworkCh,
		TriggerGatewayCh:       app.triggerFeatureControllerGatewayCh,
		TriggerIngressCh:       app.triggerFeatureControllerIngressCh,
		TriggerLoadBalancerCh:  app.triggerFeatureControllerLoadBalancerCh,
		TriggerDNSCh:           app.triggerFeatureControllerDNSCh,
		TriggerLocalStorageCh:  app.triggerFeatureControllerLocalStorageCh,
		TriggerMetricsServerCh: app.triggerFeatureControllerMetricsServerCh,
	})

	return app, nil
}

// Run starts the microcluster node and waits until it terminates.
// any non-nil customHooks override the default hooks.
func (a *App) Run(ctx context.Context, customHooks *config.Hooks) error {
	// TODO: consider improving API for overriding hooks.
	hooks := &config.Hooks{
		PostBootstrap: a.onBootstrap,
		PostJoin:      a.onPostJoin,
		PreRemove:     a.onPreRemove,
		OnStart:       a.onStart,
	}
	if customHooks != nil {
		if customHooks.PostBootstrap != nil {
			hooks.PostBootstrap = customHooks.PostBootstrap
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

	// start profiling server
	if a.profilingAddress != "" {
		log.Printf("Enable pprof endpoint at http://%s", a.profilingAddress)

		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

			if err := http.ListenAndServe(a.profilingAddress, mux); err != nil {
				log.Printf("ERROR: Failed to serve pprof endpoint: %v", err)
			}
		}()
	}

	err := a.microCluster.Start(ctx, api.New(a).Endpoints(), database.SchemaExtensions, hooks)
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
