package controllers

import (
	"context"
	"fmt"
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
		client, err := GetNewK8sClientWithRetries(ctx, c.snap)
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
	if azFound {
		log.Info("Node availability zone found", "label", azLabel)
	} else {
		log.Info("The node availability zone label is unset, skipping...")
		return nil
	}

	// k8s-dqlite expects the failure domain (availability zone) to be an uint64
	// value defined in $dbStateDir/failure-domain. Both k8s-snap Dqlite databases
	// need to be updated (k8sd and k8s-dqlite).
	failureDomain := snaputil.NodeLabelToDqliteFailureDomain(azLabel)
	log.Info("Setting failure domain", "failure domain", failureDomain, "availability zone", azLabel)
	needsRestart, err := snaputil.UpdateDqliteFailureDomain(c.snap, failureDomain)
	if err != nil {
		return fmt.Errorf("failed to update failure-domain, error: %w", err)
	}
	log.Info("Updated failure domain", "restart needed", needsRestart)

	if needsRestart {
		if err := c.snap.RestartService(ctx, "k8s-dqlite"); err != nil {
			return fmt.Errorf("failed to restart k8s-dqlite to apply failure domain: %w", err)
		}

		// TODO: k8sd restarts itself, is it safe to do so?
		if err := c.snap.RestartService(ctx, "k8sd"); err != nil {
			return fmt.Errorf("failed to restart k8s-dqlite to apply failure domain: %w", err)
		}
		// We shouldn't actually get here.
	}

	return nil
}

func (c *NodeLabelController) reconcile(ctx context.Context, node *v1.Node) error {
	log := log.FromContext(ctx)
	log.Info("reconciling node labels", "name", node.Name)

	err := c.reconcileFailureDomain(ctx, node)
	if err != nil {
		return fmt.Errorf("failed to reconcile failure domain: %w", err)
	}

	return nil
}
