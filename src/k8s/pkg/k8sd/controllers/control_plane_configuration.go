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
	// reconciledCh is used to notify that the controller has finished its reconciliation loop.
	reconciledCh chan struct{}
}

// NewControlPlaneConfigurationController creates a new controller.
// triggerCh is typically a `time.NewTicker(<duration>).C`.
func NewControlPlaneConfigurationController(snap snap.Snap, waitReady func(), triggerCh <-chan time.Time) *ControlPlaneConfigurationController {
	return &ControlPlaneConfigurationController{
		snap:         snap,
		waitReady:    waitReady,
		triggerCh:    triggerCh,
		reconciledCh: make(chan struct{}, 1),
	}
}

// Run starts the controller.
// Run accepts a context to manage the lifecycle of the controller.
// Run accepts a function that retrieves the current cluster configuration.
// Run will loop every time the trigger channel is.
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

		config, err := getClusterConfig(ctx)
		if err != nil {
			log.Error(err, "Failed to retrieve cluster configuration")
			continue
		}

		if err := c.reconcile(ctx, config); err != nil {
			log.Error(err, "Failed to reconcile control plane configuration")
		}

		select {
		case c.reconciledCh <- struct{}{}:
		default:
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
		updateArgs, deleteArgs, err := config.Datastore.ToKubeAPIServerArguments(c.snap)
		if err != nil {
			return fmt.Errorf("failed to get datastore arguments for kube-apiserver: %w", err)
		}

		argsChanged, err := snaputil.UpdateServiceArguments(c.snap, "kube-apiserver", updateArgs, deleteArgs)
		if err != nil {
			return fmt.Errorf("failed to update kube-apiserver datastore arguments: %w", err)
		}

		if certificatesChanged || argsChanged {
			if err := c.snap.RestartServices(ctx, []string{"kube-apiserver"}); err != nil {
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
			if err := c.snap.RestartServices(ctx, []string{"kube-controller-manager"}); err != nil {
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

// ReconciledCh returns the channel where the controller pushes when a reconciliation loop is finished.
func (c *ControlPlaneConfigurationController) ReconciledCh() <-chan struct{} {
	return c.reconciledCh
}
