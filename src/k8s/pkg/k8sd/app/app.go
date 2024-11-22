package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api"
	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/k8sd/controllers/csrsigning"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/microcluster/v3/client"
	"github.com/canonical/microcluster/v3/microcluster"
	"github.com/canonical/microcluster/v3/state"
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
	// DisableNodeConfigController is a bool flag to disable node config controller
	DisableNodeConfigController bool
	// DisableControlPlaneConfigController is a bool flag to disable control-plane config controller
	DisableControlPlaneConfigController bool
	// DisableUpdateNodeConfigController is a bool flag to disable update node config controller
	DisableUpdateNodeConfigController bool
	// DisableFeatureController is a bool flag to disable feature controller
	DisableFeatureController bool
	// DisableCSRSigningController is a bool flag to disable csrsigning controller.
	DisableCSRSigningController bool
}

// App is the k8sd microcluster instance.
type App struct {
	config  Config
	cluster *microcluster.MicroCluster
	client  *client.Client
	snap    snap.Snap

	// profilingAddress
	profilingAddress string

	// readyWg is used to denote that the microcluster node is now running
	readyWg sync.WaitGroup

	nodeConfigController         *controllers.NodeConfigurationController
	controlPlaneConfigController *controllers.ControlPlaneConfigurationController
	csrsigningController         *csrsigning.Controller

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
		StateDir: cfg.StateDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create microcluster app: %w", err)
	}
	client, err := cluster.LocalClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create microcluster local client: %w", err)
	}

	app := &App{
		config:           cfg,
		cluster:          cluster,
		client:           client,
		snap:             cfg.Snap,
		profilingAddress: cfg.PprofAddress,
	}
	app.readyWg.Add(1)

	if !cfg.DisableNodeConfigController {
		app.nodeConfigController = controllers.NewNodeConfigurationController(
			cfg.Snap,
			app.readyWg.Wait,
		)
	} else {
		log.L().Info("node-config-controller disabled via config")
	}

	if !cfg.DisableControlPlaneConfigController {
		app.controlPlaneConfigController = controllers.NewControlPlaneConfigurationController(
			cfg.Snap,
			app.readyWg.Wait,
			time.NewTicker(10*time.Second).C,
		)
	} else {
		log.L().Info("control-plane-config-controller disabled via config")
	}

	app.triggerUpdateNodeConfigControllerCh = make(chan struct{}, 1)

	if !cfg.DisableUpdateNodeConfigController {
		app.updateNodeConfigController = controllers.NewUpdateNodeConfigurationController(
			cfg.Snap,
			app.readyWg.Wait,
			app.triggerUpdateNodeConfigControllerCh,
		)
	} else {
		log.L().Info("update-node-config-controller disabled via config")
	}

	app.triggerFeatureControllerNetworkCh = make(chan struct{}, 1)
	app.triggerFeatureControllerGatewayCh = make(chan struct{}, 1)
	app.triggerFeatureControllerIngressCh = make(chan struct{}, 1)
	app.triggerFeatureControllerLoadBalancerCh = make(chan struct{}, 1)
	app.triggerFeatureControllerLocalStorageCh = make(chan struct{}, 1)
	app.triggerFeatureControllerMetricsServerCh = make(chan struct{}, 1)
	app.triggerFeatureControllerDNSCh = make(chan struct{}, 1)

	if !cfg.DisableFeatureController {
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
	} else {
		log.L().Info("feature-controller disabled via config")
	}

	if !cfg.DisableCSRSigningController {
		app.csrsigningController = csrsigning.New(csrsigning.Options{
			Snap:           cfg.Snap,
			WaitReady:      app.readyWg.Wait,
			LeaderElection: true,
		})
	} else {
		log.L().Info("csrsigning-controller disabled via config")
	}

	return app, nil
}

// Run starts the microcluster node and waits until it terminates.
// any non-nil customHooks override the default hooks.
func (a *App) Run(ctx context.Context, customHooks *state.Hooks) error {
	// TODO: consider improving API for overriding hooks.
	hooks := &state.Hooks{
		PreInit:       a.onPreInit,
		PostBootstrap: a.onBootstrap,
		PostJoin:      a.onPostJoin,
		PreRemove:     a.onPreRemove,
		OnStart:       a.onStart,
	}
	if customHooks != nil {
		if customHooks.PreInit != nil {
			hooks.PreInit = customHooks.PreInit
		}
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

	log := log.FromContext(ctx)

	// start profiling server
	if a.profilingAddress != "" {
		log.WithValues("address", fmt.Sprintf("http://%s", a.profilingAddress)).Info("Enable pprof endpoint")

		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

			if err := http.ListenAndServe(a.profilingAddress, mux); err != nil {
				log.Error(err, "Failed to serve pprof endpoint")
			}
		}()
	}

	err := a.cluster.Start(ctx, microcluster.DaemonArgs{
		Version:          string(apiv1.K8sdAPIVersion),
		Verbose:          a.config.Verbose,
		Debug:            a.config.Debug,
		Hooks:            hooks,
		ExtensionServers: api.New(ctx, a),
		ExtensionsSchema: database.SchemaExtensions,
	})
	if err != nil {
		return fmt.Errorf("failed to run microcluster: %w", err)
	}
	return nil
}

// markNodeReady will decrement the readyWg counter to signal that the node is ready.
// The node is ready if:
// - the microcluster database is accessible
// - the kubernetes endpoint is reachable.
func (a *App) markNodeReady(ctx context.Context, s state.State) error {
	log := log.FromContext(ctx).WithValues("startup", "waitForReady")

	// wait for the database to be open
	log.V(1).Info("Waiting for database to be open")
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		return s.Database().IsOpen(ctx) == nil, nil
	}); err != nil {
		return fmt.Errorf("failed to wait for database to be open: %w", err)
	}

	// check kubernetes endpoint
	log.V(1).Info("Waiting for kubernetes endpoint")
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		client, err := a.snap.KubernetesNodeClient("")
		if err != nil {
			return false, nil
		}
		if err := client.CheckKubernetesEndpoint(ctx); err != nil {
			return false, nil
		}
		return true, nil
	}); err != nil {
		return fmt.Errorf("failed to wait for kubernetes endpoint: %w", err)
	}

	log.V(1).Info("Marking node as ready")
	a.readyWg.Done()
	return nil
}
