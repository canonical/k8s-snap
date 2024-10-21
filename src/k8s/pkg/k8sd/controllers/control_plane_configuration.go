package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
)

// ControlPlaneConfigurationController watches for changes in the cluster configuration
// and applies them on the control plane services.
type ControlPlaneConfigurationController struct {
	snap      snap.Snap
	waitReady func()
	triggerCh <-chan time.Time
}

// NewControlPlaneConfigurationController creates a new controller.
// triggerCh is typically a `time.NewTicker(<duration>).C`
func NewControlPlaneConfigurationController(snap snap.Snap, waitReady func(), triggerCh <-chan time.Time) *ControlPlaneConfigurationController {
	return &ControlPlaneConfigurationController{
		snap:      snap,
		waitReady: waitReady,
		triggerCh: triggerCh,
	}
}

// Run starts the controller.
// Run accepts a context to manage the lifecycle of the controller.
// Run accepts a function that retrieves the current cluster configuration.
// Run will loop every time the trigger channel is
func (c *ControlPlaneConfigurationController) Run(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error)) {
	c.waitReady()

	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "control-plane-configuration"))
	log := log.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.triggerCh:
		}

		if isWorker, err := snaputil.IsWorker(c.snap); err != nil {
			log.Error(err, "Failed to check if running on a worker node")
			continue
		} else if isWorker {
			log.Info("Stopping on worker node")
			return
		}

		config, err := getClusterConfig(ctx)
		if err != nil {
			log.Error(err, "Failed to retrieve cluster configuration")
			continue
		}

		if err := c.reconcile(ctx, config); err != nil {
			log.Error(err, "Failed to reconcile control plane configuration")
		}
	}
}

func (c *ControlPlaneConfigurationController) reconcile(ctx context.Context, config types.ClusterConfig) error {
	// kube-apiserver: external datastore
	if config.Datastore.GetType() == "external" {
		// certificates
		certificatesChanged, err := setup.EnsureExtDatastorePKI(c.snap, &pki.ExternalDatastorePKI{
			DatastoreCACert:     config.Datastore.GetExternalCACert(),
			DatastoreClientCert: config.Datastore.GetExternalClientCert(),
			DatastoreClientKey:  config.Datastore.GetExternalClientKey(),
		})
		if err != nil {
			return fmt.Errorf("failed to reconcile external datastore certificates: %w", err)
		}

		// kube-apiserver arguments
		updateArgs, deleteArgs := config.Datastore.ToKubeAPIServerArguments(c.snap)
		argsChanged, err := snaputil.UpdateServiceArguments(c.snap, "kube-apiserver", updateArgs, deleteArgs)
		if err != nil {
			return fmt.Errorf("failed to update kube-apiserver datastore arguments: %w", err)
		}

		if certificatesChanged || argsChanged {
			if err := c.snap.RestartService(ctx, "kube-apiserver"); err != nil {
				return fmt.Errorf("failed to restart kube-apiserver to apply configuration: %w", err)
			}
		}
	}

	// kube-controller-manager: cloud-provider
	if v := config.Kubelet.CloudProvider; v != nil {
		mustRestart, err := snaputil.UpdateServiceArguments(c.snap, "kube-controller-manager", map[string]string{"--cloud-provider": *v}, nil)
		if err != nil {
			return fmt.Errorf("failed to update kube-controller-manager arguments: %w", err)
		}

		if mustRestart {
			if err := c.snap.RestartService(ctx, "kube-controller-manager"); err != nil {
				return fmt.Errorf("failed to restart kube-controller-manager to apply configuration: %w", err)
			}
		}
	}

	// snapd
	if meta, _, err := snapdconfig.ParseMeta(ctx, c.snap); err == nil && meta.Orb != "none" {
		if err := snapdconfig.SetSnapdFromK8sd(ctx, config.ToUserFacing(), c.snap); err != nil {
			log.FromContext(ctx).Error(err, "Failed to update snapd configuration")
		}
	}

	return nil
}
