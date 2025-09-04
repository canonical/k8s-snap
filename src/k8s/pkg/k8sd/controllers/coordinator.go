package controllers

import (
	"context"
	"fmt"
	"time"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/controllers/csrsigning"
	"github.com/canonical/k8s/pkg/k8sd/controllers/upgrade"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// Coordinator is responsible for managing the lifecycle of controllers.
type Coordinator struct {
	snap      snap.Snap
	waitReady func()

	upgradeControllerOptions    UpgradeControllerOptions
	csrSigningControllerOptions CSRSigningControllerOptions
}

type UpgradeControllerOptions struct {
	upgrade.ControllerOptions
	Disable bool
}

type CSRSigningControllerOptions struct {
	Disable bool
}

// NewCoordinator creates a new Coordinator instance.
func NewCoordinator(
	snap snap.Snap,
	waitReady func(),
	upgradeControllerOptions UpgradeControllerOptions,
	csrSigningControllerOptions CSRSigningControllerOptions,
) *Coordinator {
	return &Coordinator{
		snap:                        snap,
		waitReady:                   waitReady,
		upgradeControllerOptions:    upgradeControllerOptions,
		csrSigningControllerOptions: csrSigningControllerOptions,
	}
}

// Run creates a manager, setup the controllers with the manager and starts the manager.
func (c *Coordinator) Run(
	ctx context.Context,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
) error {
	logger := log.FromContext(ctx).WithName("controller-coordinator")

	isWorker, err := snaputil.IsWorker(c.snap)
	if err != nil {
		return fmt.Errorf("failed to determine if snap is running as worker: %w", err)
	}
	if isWorker {
		logger.Info("Skipping controller coordinator on worker node")
		return nil
	}

	ctrllog.SetLogger(logger)
	ctx = log.NewContext(ctx, logger)

	readyCh := make(chan struct{})
	go func() {
		defer close(readyCh)
		c.waitReady()
	}()
	select {
	case <-readyCh:
	case <-ctx.Done():
		return fmt.Errorf("failed to wait for Microcluster to be ready: %w", ctx.Err())
	}

	config, err := c.getRESTConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes REST config: %w", err)
	}

	scheme, err := kubernetes.NewScheme()
	if err != nil {
		return fmt.Errorf("failed to create scheme: %w", err)
	}

	mgr, err := manager.New(config, manager.Options{
		Scheme:                  scheme,
		Logger:                  logger,
		LeaderElection:          true,
		LeaderElectionID:        "oy6981cu.controller-coordinator",
		LeaderElectionNamespace: "kube-system",
		BaseContext:             func() context.Context { return ctx },
		Cache: cache.Options{
			SyncPeriod: utils.Pointer(10 * time.Minute),
		},
		Metrics: server.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	if err := c.setupControllers(ctx, getClusterConfig, mgr); err != nil {
		return fmt.Errorf("failed to setup controllers: %w", err)
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start manager: %w", err)
	}

	return nil
}

func (c *Coordinator) setupControllers(
	ctx context.Context,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
	mgr manager.Manager,
) error {
	if err := c.setupUpgradeController(ctx, getClusterConfig, mgr); err != nil {
		return fmt.Errorf("failed to setup upgrade controller: %w", err)
	}

	if err := c.setupCSRSigningController(getClusterConfig, mgr); err != nil {
		return fmt.Errorf("failed to setup CSR signing controller: %w", err)
	}

	return nil
}

func (c *Coordinator) setupUpgradeController(
	ctx context.Context,
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
	mgr manager.Manager,
) error {
	logger := mgr.GetLogger()

	if c.upgradeControllerOptions.Disable {
		logger.Info("Upgrade controller is disabled. Skipping setup.")
		return nil
	}

	clusterConfig, err := getClusterConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve cluster configuration: %w", err)
	}

	if featureUpgradesDisabled(clusterConfig) {
		logger.Info("Feature upgrades are disabled. Skipping upgrade controller.")
		return nil
	}

	upgradeController := upgrade.NewController(
		logger,
		mgr.GetClient(),
		c.upgradeControllerOptions.ControllerOptions,
	)

	if err := upgradeController.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup upgrade controller with manager: %w", err)
	}

	return nil
}

func (c *Coordinator) setupCSRSigningController(
	getClusterConfig func(context.Context) (types.ClusterConfig, error),
	mgr manager.Manager,
) error {
	logger := mgr.GetLogger()

	if c.csrSigningControllerOptions.Disable {
		logger.Info("CSR signing controller is disabled. Skipping setup.")
		return nil
	}

	csrsigningController := csrsigning.NewController(
		logger,
		mgr.GetClient(),
		getClusterConfig,
	)

	if err := csrsigningController.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup csrsigning controller with manager: %w", err)
	}

	return nil
}

// featureUpgradesDisabled checks if feature upgrades are disabled in the cluster configuration.
func featureUpgradesDisabled(clusterConfig types.ClusterConfig) bool {
	_, ok := clusterConfig.Annotations.Get(apiv1_annotations.AnnotationDisableSeparateFeatureUpgrades)
	return ok
}

func (c *Coordinator) getRESTConfig(ctx context.Context) (*rest.Config, error) {
	for {
		client, err := c.snap.KubernetesClient("")
		if err == nil {
			return client.RESTConfig(), nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}
