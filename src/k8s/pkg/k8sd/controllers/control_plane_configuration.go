package controllers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
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

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.triggerCh:
		}

		if isWorker, err := snaputil.IsWorker(c.snap); err != nil {
			log.Println(fmt.Errorf("failed to check if this is a worker node: %w", err))
			continue
		} else if isWorker {
			log.Println("Stopping control plane controller as this is a worker node")
			return
		}

		config, err := getClusterConfig(ctx)
		if err != nil {
			log.Println(fmt.Errorf("failed to retrieve cluster config: %w", err))
			continue
		}

		if err := c.reconcile(ctx, config); err != nil {
			log.Println(fmt.Errorf("failed to reconcile control plane configuration: %w", err))
		}
	}
}

func (c *ControlPlaneConfigurationController) reconcile(ctx context.Context, config types.ClusterConfig) error {
	// kube-apiserver: external datastore
	switch config.Datastore.GetType() {
	case "external":
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

	return nil
}
