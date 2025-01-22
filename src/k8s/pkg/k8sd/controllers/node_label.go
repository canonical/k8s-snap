package controllers

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	v1 "k8s.io/api/core/v1"
)

type NodeLabelController struct {
	snap      snap.Snap
	waitReady func()
}

func NewNodeLabelController(snap snap.Snap, waitReady func()) *NodeLabelController {
	return &NodeLabelController{
		snap:      snap,
		waitReady: waitReady,
	}
}

func (c *NodeLabelController) Run(ctx context.Context) {
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "node-configuration"))
	log := log.FromContext(ctx)

	log.Info("Waiting for node to be ready")
	// wait for microcluster node to be ready
	c.waitReady()

	hostname := c.snap.Hostname()
	log.Info("Starting node label controller", "hostname", hostname)

	for {
		client, err := getNewK8sClientWithRetries(ctx, c.snap)
		if err != nil {
			log.Error(err, "Failed to create a Kubernetes client")
		}

		if err := client.WatchNode(
			ctx, hostname, func(node *v1.Node) error { return c.reconcile(ctx, node) }); err != nil {
			// The watch may fail during bootstrap or service start-up.
			log.WithValues("node name", hostname).Error(err, "Failed to watch node")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (c *NodeLabelController) reconcileFailureDomain(ctx context.Context, node *v1.Node) error {
	log := log.FromContext(ctx)

	azLabel, azFound := node.Labels["topology.kubernetes.io/zone"]
	var failureDomain uint64
	if azFound && azLabel != "" {
		log.Info("Node availability zone found", "label", azLabel)
		// k8s-dqlite expects the failure domain (availability zone) to be an uint64
		// value defined in $dbStateDir/failure-domain. Both k8s-snap Dqlite databases
		// need to be updated (k8sd and k8s-dqlite).
		failureDomain = snaputil.NodeLabelToDqliteFailureDomain(azLabel)
	} else {
		log.Info("The node availability zone label is unset, clearing failure domain")
		failureDomain = 0
	}

	log.Info("Setting failure domain", "failure domain", failureDomain, "availability zone", azLabel)
	err := c.updateDqliteFailureDomain(ctx, c.snap, failureDomain)
	if err != nil {
		return fmt.Errorf("failed to update failure-domain, error: %w", err)
	}

	return nil
}

func (c *NodeLabelController) updateDqliteFailureDomain(ctx context.Context, snap snap.Snap, failureDomain uint64) error {
	log := log.FromContext(ctx)

	// We need to update both k8s-snap Dqlite databases (k8sd and k8s-dqlite).
	k8sDqliteStateDir := snap.K8sDqliteStateDir()
	k8sdDbStateDir := filepath.Join(snap.K8sdStateDir(), "database")

	log.Info("Updating k8s-dqlite failure domain", "failure domain", failureDomain)
	modified, err := snaputil.UpdateDqliteFailureDomain(failureDomain, k8sDqliteStateDir)
	if err != nil {
		return err
	}
	log.Info("Updated k8s-dqlite failure domain", "restart needed", modified)

	if modified {
		if err = c.snap.RestartService(ctx, "k8s-dqlite"); err != nil {
			return fmt.Errorf("failed to restart k8s-dqlite to apply failure domain: %w", err)
		}
	}

	log.Info("Updating k8sd failure domain", "failure domain", failureDomain)
	modified, err = snaputil.UpdateDqliteFailureDomain(failureDomain, k8sdDbStateDir)
	if err != nil {
		return err
	}
	log.Info("Updated k8sd failure domain", "restart needed", modified)
	// TODO: use Microcluster API once it becomes available. This should
	// prevent a service restart, at the moment k8sd needs to restart itself.
	if modified {
		if err := c.snap.RestartService(ctx, "k8sd"); err != nil {
			return fmt.Errorf("failed to restart k8sd to apply failure domain: %w", err)
		}
		// We shouldn't actually get here.
	}

	return nil
}

func (c *NodeLabelController) reconcile(ctx context.Context, node *v1.Node) error {
	log := log.FromContext(ctx)
	log.Info("reconciling node labels", "name", node.Name)

	if err := c.reconcileFailureDomain(ctx, node); err != nil {
		return fmt.Errorf("failed to reconcile failure domain: %w", err)
	}

	return nil
}
