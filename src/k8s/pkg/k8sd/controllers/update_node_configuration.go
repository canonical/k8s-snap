package controllers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
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

func (c *UpdateNodeConfigurationController) retryNewK8sClient(ctx context.Context) (*kubernetes.Client, error) {
	for {
		client, err := c.snap.KubernetesClient("kube-system")
		if err == nil {
			return client, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

// Run starts the controller.
// Run accepts a context to manage the lifecycle of the controller.
// Run accepts a function that retrieves the current cluster configuration.
// Run will loop everytime the TriggerCh is triggered.
func (c *UpdateNodeConfigurationController) Run(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error)) {
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
			log.Println("Stopping UpdateClusterConfig controller as this is a worker node")
			return
		}

		config, err := getClusterConfig(ctx)
		if err != nil {
			log.Println(fmt.Errorf("failed to retrieve cluster config: %w", err))
			continue
		}

		client, err := c.retryNewK8sClient(ctx)
		if err != nil {
			log.Println(fmt.Errorf("failed to create a Kubernetes client: %w", err))
		}

		if err := c.reconcile(ctx, client, config); err != nil {
			log.Println(fmt.Errorf("failed to reconcile cluster configuration: %w", err))
		}

		// notify downstream that the reconciliation loop is done.
		select {
		case c.reconciledCh <- struct{}{}:
		default:
		}
	}
}

func (c *UpdateNodeConfigurationController) reconcile(ctx context.Context, client *kubernetes.Client, config types.ClusterConfig) error {
	keyPEM := config.Certificates.GetK8sdPrivateKey()
	key, err := pki.LoadRSAPrivateKey(keyPEM)
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
