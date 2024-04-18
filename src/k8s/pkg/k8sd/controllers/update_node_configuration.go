package controllers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/k8s"
)

// UpdateNodeConfigurationController asynchronously performs updates of the cluster config.
// An updates is triggered by sending to the TriggerCh.
type UpdateNodeConfigurationController struct {
	snap         snap.Snap
	waitReady    func()
	newK8sClient func() (*k8s.Client, error)
	// TriggerCh is used to trigger config updates.
	TriggerCh chan struct{}
	// ReconciledCh is used to indicate that a reconcilation loop has finished.
	ReconciledCh chan struct{}
}

// NewUpdateNodeConfigurationController creates a new controller.
func NewUpdateNodeConfigurationController(snap snap.Snap, waitReady func(), newK8sClient func() (*k8s.Client, error)) *UpdateNodeConfigurationController {
	return &UpdateNodeConfigurationController{
		snap:         snap,
		waitReady:    waitReady,
		newK8sClient: newK8sClient,
		TriggerCh:    make(chan struct{}, 1),
		ReconciledCh: make(chan struct{}, 1),
	}
}

func (c *UpdateNodeConfigurationController) retryNewK8sClient(ctx context.Context) (*k8s.Client, error) {
	for {
		client, err := c.newK8sClient()
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
		client, err := c.retryNewK8sClient(ctx)
		if err != nil {
			log.Println(fmt.Errorf("failed to create a Kubernetes client: %w", err))
		}

		select {
		case <-ctx.Done():
			return
		case <-c.TriggerCh:
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

		if err := c.reconcile(ctx, client, config); err != nil {
			log.Println(fmt.Errorf("failed to reconcile cluster configuration: %w", err))
		}

		select {
		case c.ReconciledCh <- struct{}{}:
		default:
		}
	}
}

func (c *UpdateNodeConfigurationController) reconcile(ctx context.Context, client *k8s.Client, config types.ClusterConfig) error {
	cmData, err := config.Kubelet.ToConfigMap(nil)
	if err != nil {
		return fmt.Errorf("failed to format kubelet configmap data: %w", err)
	}
	if _, err := client.UpdateConfigMap(ctx, "kube-system", "k8sd-config", cmData); err != nil {
		return fmt.Errorf("failed to update node config: %w", err)
	}

	return nil
}
