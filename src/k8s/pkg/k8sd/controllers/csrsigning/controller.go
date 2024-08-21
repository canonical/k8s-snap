package csrsigning

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
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

func (c *Controller) Run(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error)) error {
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithName("csrsigning"))

	// TODO(neoaggelos): This should be moved to init() or some other initialization step
	ctrllog.SetLogger(log.FromContext(ctx))

	c.waitReady()

	config, err := c.getRESTConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes REST config: %w", err)
	}

	// TODO(neoaggelos): In case of more controllers, a single manager object should be created
	// and passed here as configuration.
	mgr, err := manager.New(config, manager.Options{
		Logger:                  log.FromContext(ctx),
		LeaderElection:          c.leaderElection,
		LeaderElectionID:        "a27980c4.k8sd-csrsigning-controller",
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

	if err := (&csrSigningReconciler{
		Manager:            mgr,
		Logger:             mgr.GetLogger(),
		Client:             mgr.GetClient(),
		managedSignerNames: managedSignerNames,

		getClusterConfig:     getClusterConfig,
		reconcileAutoApprove: reconcileAutoApprove,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup csrsigning controller: %w", err)
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("controller manager failed: %w", err)
	}

	return nil
}
