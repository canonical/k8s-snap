package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	defaultFeatureControllerReadyTimeout     = 30 * time.Second
	defaultFeatureControllerReconcileTimeout = 30 * time.Second
)

type Controller struct {
	snap      snap.Snap
	waitReady func()

	leaderElection bool
}

type Options struct {
	Snap      snap.Snap
	WaitReady func()

	LeaderElection bool
}

func New(opts Options) *Controller {
	return &Controller{
		snap:      opts.Snap,
		waitReady: opts.WaitReady,

		leaderElection: opts.LeaderElection,
	}
}

func (c *Controller) getRESTConfig(ctx context.Context) (*rest.Config, error) {
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

func (c *Controller) Run(
	ctx context.Context,
	getState func() state.State,
	featureControllerReadyCh <-chan struct{},
	notifyFeatureController func(),
	reconciledNetworkCh <-chan struct{},
	reconciledGatewayCh <-chan struct{},
	reconciledIngressCh <-chan struct{},
	reconciledDNSCh <-chan struct{},
	reconciledLoadBalancerCh <-chan struct{},
	reconciledLocalStorageCh <-chan struct{},
	reconciledMetricsServerCh <-chan struct{},
) error {
	logger := log.FromContext(ctx).WithName("upgrade-controller")
	ctx = log.NewContext(ctx, logger)

	c.waitReady()

	config, err := c.getRESTConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes REST config: %w", err)
	}

	// TODO(Hue): (KU-3216) use a single manager for upgrade and csrsigning controllers.
	mgr, err := manager.New(config, manager.Options{
		Logger:                  logger,
		LeaderElection:          c.leaderElection,
		LeaderElectionID:        "a27980c4.k8sd-upgrade-controller",
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
		return fmt.Errorf("failed to create controller manager: %w", err)
	}

	if err := (&upgradeReconciler{
		getState:                          getState,
		snap:                              c.snap,
		featureControllerReadyTimeout:     defaultFeatureControllerReadyTimeout,
		featureControllerReconcileTimeout: defaultFeatureControllerReconcileTimeout,
		featureControllerReadyCh:          featureControllerReadyCh,
		notifyFeatureController:           notifyFeatureController,
		featureToReconciledCh: map[string]<-chan struct{}{
			"network":        reconciledNetworkCh,
			"gateway":        reconciledGatewayCh,
			"ingress":        reconciledIngressCh,
			"dns":            reconciledDNSCh,
			"load-balancer":  reconciledLoadBalancerCh,
			"local-storage":  reconciledLocalStorageCh,
			"metrics-server": reconciledMetricsServerCh,
		},

		Manager: mgr,
		Logger:  mgr.GetLogger(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup upgrade controller: %w", err)
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("controller manager failed: %w", err)
	}

	return nil
}
