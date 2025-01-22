package controllers

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
)

// UpdateNodeConfigurationController asynchronously performs updates of the cluster config.
// A new reconcile loop is triggered by pushing to the triggerCh channel.
type UpdateNodeConfigurationController struct {
	snap      snap.Snap
	waitReady func()

	// triggerCh is used to trigger config updates on the controller.
	triggerCh <-chan struct{}
	// reconciledCh is used to notify that the controller has finished its reconciliation loop.
	reconciledCh chan struct{}
}

// NewUpdateNodeConfigurationController creates a new controller.
func NewUpdateNodeConfigurationController(snap snap.Snap, waitReady func(), triggerCh <-chan struct{}) *UpdateNodeConfigurationController {
	return &UpdateNodeConfigurationController{
		snap:      snap,
		waitReady: waitReady,

		triggerCh:    triggerCh,
		reconciledCh: make(chan struct{}, 1),
	}
}

// Run starts the controller.
// Run accepts a context to manage the lifecycle of the controller.
// Run accepts a function that retrieves the current cluster configuration.
// Run will loop everytime the TriggerCh is triggered.
func (c *UpdateNodeConfigurationController) Run(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error)) {
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "update-node-configuration"))
	log := log.FromContext(ctx)

	log.V(1).Info("Waiting for node to be ready")
	c.waitReady()
	log.V(1).Info("Starting update node configuration controller")

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

		client, err := getNewK8sClientWithRetries(ctx, c.snap)
		if err != nil {
			log.Error(err, "Failed to create a Kubernetes client")
		}

		if err := c.reconcile(ctx, client, config); err != nil {
			log.Error(err, "Failed to reconcile cluster configuration")
		}

		// notify downstream that the reconciliation loop is done.
		select {
		case c.reconciledCh <- struct{}{}:
		default:
		}
	}
}

func (c *UpdateNodeConfigurationController) reconcile(ctx context.Context, client *kubernetes.Client, config types.ClusterConfig) error {
	log := log.FromContext(ctx)
	log.V(1).Info("Reconciling node configuration")

	keyPEM := config.Certificates.GetK8sdPrivateKey()
	key, err := pkiutil.LoadRSAPrivateKey(keyPEM)
	if err != nil && keyPEM != "" {
		return fmt.Errorf("failed to load cluster RSA key: %w", err)
	}

	cmData, err := config.Kubelet.ToConfigMap(key)
	if err != nil {
		return fmt.Errorf("failed to format kubelet configmap data: %w", err)
	}
	if _, err := client.UpdateConfigMap(ctx, "kube-system", "k8sd-config", cmData); err != nil {
		return fmt.Errorf("failed to update node config: %w", err)
	}

	return nil
}

// ReconciledCh returns the channel where the controller pushes when a reconciliation loop is finished.
func (c *UpdateNodeConfigurationController) ReconciledCh() <-chan struct{} {
	return c.reconciledCh
}
