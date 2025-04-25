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
	// reconciledCh is used to notify that the controller has finished its reconciliation loop.
	reconciledCh chan struct{}
}

func NewNodeLabelController(snap snap.Snap, waitReady func()) *NodeLabelController {
	return &NodeLabelController{
		snap:         snap,
		waitReady:    waitReady,
		reconciledCh: make(chan struct{}, 1),
	}
}

func (c *NodeLabelController) Run(ctx context.Context, nodeName string) {
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "node-configuration"))
	log := log.FromContext(ctx)

	log.Info("Waiting for node to be ready")
	// wait for microcluster node to be ready
	c.waitReady()

	log.Info("Starting node label controller", "nodeName", nodeName)

	for {
		client, err := getNewK8sClientWithRetries(ctx, c.snap, false)
		if err != nil {
			log.Error(err, "Failed to create a Kubernetes client")
		}

		if err := client.WatchNode(
			ctx, nodeName, func(node *v1.Node) error {
				err := c.reconcile(ctx, node)
				c.notifyReconciled()
				return err
			}); err != nil {
			// The watch may fail during bootstrap or service start-up.
			log.WithValues("node name", nodeName).Error(err, "Failed to watch node")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (c *NodeLabelController) reconcileFailureDomain(ctx context.Context, node *v1.Node) error {
	azLabel, azFound := node.Labels["topology.kubernetes.io/zone"]
	var failureDomain uint64
	if azFound && azLabel != "" {
		// k8s-dqlite expects the failure domain (availability zone) to be an uint64
		// value defined in $dbStateDir/failure-domain. Both k8s-snap Dqlite databases
		// need to be updated (k8sd and k8s-dqlite).
		failureDomain = snaputil.NodeLabelToDqliteFailureDomain(azLabel)
	} else {
		failureDomain = 0
	}

	if err := c.updateDqliteFailureDomain(ctx, failureDomain, azLabel); err != nil {
		return fmt.Errorf("failed to update failure-domain, error: %w", err)
	}

	return nil
}

func (c *NodeLabelController) updateDqliteFailureDomain(ctx context.Context, failureDomain uint64, availabilityZone string) error {
	log := log.FromContext(ctx)

	// We need to update both k8s-snap Dqlite databases (k8sd and k8s-dqlite).
	k8sDqliteStateDir := c.snap.K8sDqliteStateDir()
	k8sdDbStateDir := filepath.Join(c.snap.K8sdStateDir(), "database")

	modified, err := snaputil.UpdateDqliteFailureDomain(failureDomain, k8sDqliteStateDir)
	if err != nil {
		return err
	}

	if modified {
		log.Info("Updated k8s-dqlite failure domain", "failure domain", failureDomain, "availability zone", availabilityZone)
		if err = c.snap.RestartServices(ctx, []string{"k8s-dqlite"}); err != nil {
			return fmt.Errorf("failed to restart k8s-dqlite to apply failure domain: %w", err)
		}
	}

	modified, err = snaputil.UpdateDqliteFailureDomain(failureDomain, k8sdDbStateDir)
	if err != nil {
		return err
	}

	// TODO: use Microcluster API once it becomes available. This should
	// prevent a service restart, at the moment k8sd needs to restart itself.
	if modified {
		log.Info("Updated k8sd failure domain", "failure domain", failureDomain, "availability zone", availabilityZone)
		if err := c.snap.RestartServices(ctx, []string{"k8sd"}); err != nil {
			return fmt.Errorf("failed to restart k8sd to apply failure domain: %w", err)
		}
		// We shouldn't actually get here.
	}

	return nil
}

func (c *NodeLabelController) reconcile(ctx context.Context, node *v1.Node) error {
	if err := c.reconcileFailureDomain(ctx, node); err != nil {
		return fmt.Errorf("failed to reconcile failure domain: %w", err)
	}

	return nil
}

// ReconciledCh returns the channel where the controller pushes when a reconciliation loop is finished.
func (c *NodeLabelController) ReconciledCh() <-chan struct{} {
	return c.reconciledCh
}

func (c *NodeLabelController) notifyReconciled() {
	select {
	case c.reconciledCh <- struct{}{}:
	default:
	}
}
