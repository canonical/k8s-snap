package upgrade

import (
	"context"
	"fmt"
	"time"

	upgradesv1alpha "github.com/canonical/k8s/pkg/k8sd/crds/upgrades/v1alpha"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconcile implements the Reconciler interface and wraps the reconcile method.
// Reconcile ensures that the reconciliation is requeued unless the reconciled resource is not found.
func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := c.logger.WithValues("scope", "reconcile wrapper")

	res, err := c.reconcile(ctx, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Upgrade %q not found, ignoring.", req.NamespacedName))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to reconcile: %w", err)
	}

	bareResult := res == ctrl.Result{}
	if bareResult {
		res = ctrl.Result{RequeueAfter: 5 * time.Minute}
	}

	return res, nil
}

// reconcile is the main reconciliation loop for the upgrade controller.
func (c *Controller) reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var upgrade upgradesv1alpha.Upgrade
	if err := c.client.Get(ctx, req.NamespacedName, &upgrade); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get upgrade %q: %w", req.NamespacedName, err)
	}

	c.logger.WithValues("upgrade", upgrade.Name, "phase", upgrade.Status.Phase).Info("Reconciling upgrade.")

	switch {
	case upgrade.Status.Phase == upgradesv1alpha.UpgradePhaseNodeUpgrade:
		return c.reconcileNodeUpgrade(ctx, &upgrade)
	case upgrade.Status.Phase == upgradesv1alpha.UpgradePhaseFeatureUpgrade:
		return c.reconcileFeatureUpgrade(ctx, &upgrade)
	}

	return ctrl.Result{}, nil
}

// reconcileNodeUpgrade checks if all nodes have been upgraded.
// If so, it transitions to the feature upgrade phase and notifies the feature controller.
func (c *Controller) reconcileNodeUpgrade(ctx context.Context, upgrade *upgradesv1alpha.Upgrade) (ctrl.Result, error) {
	log := c.logger.WithValues("upgrade", upgrade.Name, "step", "node-upgrade")
	log.Info("Checking if all nodes have been upgraded.")

	allNodesUpgraded, err := c.allNodesUpgraded(ctx, upgrade.Status.UpgradedNodes)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to check if all nodes have been upgraded: %w", err)
	} else if !allNodesUpgraded {
		return ctrl.Result{}, nil
	}

	log.Info("All nodes have been upgraded.")

	if err := c.transitionTo(ctx, upgrade, upgradesv1alpha.UpgradePhaseFeatureUpgrade); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to transition to %q phase: %w", upgradesv1alpha.UpgradePhaseFeatureUpgrade, err)
	}

	log.Info(fmt.Sprintf("Transitioned to %q phase.", upgradesv1alpha.UpgradePhaseFeatureUpgrade))
	return ctrl.Result{}, nil
}

// reconcileFeatureUpgrade triggers feature controllers to reconcile
// and waits for them to finish.
func (c *Controller) reconcileFeatureUpgrade(ctx context.Context, upgrade *upgradesv1alpha.Upgrade) (ctrl.Result, error) {
	log := c.logger.WithValues("upgrade", upgrade.Name, "step", "feature-upgrade")

	log.Info("Waiting for feature controllers to be ready.")
	select {
	case <-c.featureControllerReadyCh:
	case <-time.After(c.featureControllerReadyTimeout):
		return ctrl.Result{}, fmt.Errorf("timed out waiting for feature controllers to be ready")
	}

	log.Info("Waiting for feature controllers to reconcile.")
	if err := c.waitForFeatureReconciliations(ctx, log); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to wait for feature reconciliations: %w", err)
	}

	log.Info("All feature have reconciled. Transitioning to completed phase.")
	if err := c.transitionTo(ctx, upgrade, upgradesv1alpha.UpgradePhaseCompleted); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to transition to %q phase: %w", upgradesv1alpha.UpgradePhaseCompleted, err)
	}

	log.Info(fmt.Sprintf("Transitioned to %q phase.", upgradesv1alpha.UpgradePhaseCompleted))
	return ctrl.Result{}, nil
}

// allNodesUpgraded checks if all nodes in the cluster have been upgraded.
func (c *Controller) allNodesUpgraded(ctx context.Context, upgradedNodes []string) (bool, error) {
	log := c.logger.WithValues("step", "all-nodes-upgraded")

	leader, err := c.getState().Leader()
	if err != nil {
		return false, fmt.Errorf("failed to get leader client: %w", err)
	}

	clusterMembers, err := leader.GetClusterMembers(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get cluster members: %w", err)
	}

	if len(upgradedNodes) < len(clusterMembers) {
		log.Info("Not all nodes have been upgraded.", "clusterMembers", len(clusterMembers), "upgradedNodes", len(upgradedNodes))
		return false, nil
	}

	upgradedNodesMap := make(map[string]struct{})
	for _, n := range upgradedNodes {
		upgradedNodesMap[n] = struct{}{}
	}

	// NOTE(Hue): We only need to make sure all the nodes that are part of the cluster
	// are upgraded. Don't care about upgraded nodes that are not part of the cluster.
	// Maybe they've left, are removed, etc.
	for _, member := range clusterMembers {
		if _, ok := upgradedNodesMap[member.Name]; !ok {
			log.Info(fmt.Sprintf("Cluster member %q is not upgraded", member.Name), "member_name", member.Name)
			return false, nil
		}
	}

	return true, nil
}

func (c *Controller) waitForFeatureReconciliations(ctx context.Context, log logr.Logger) error {
	for name, ch := range c.featureToReconciledCh {
		if err := c.triggerFeature(name); err != nil {
			return fmt.Errorf("failed to trigger feature %q: %w", name, err)
		}

		timeout := time.After(c.featureControllerReconcileTimeout)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			// TODO(Hue): (KU-3227) Do something about failed feature reconciliations.
			return fmt.Errorf("timed out waiting for feature %q to get reconciled", name)
		case <-ch:
			log.Info(fmt.Sprintf("feature %q have reconciled.", name))
		}
	}

	return nil
}

func (c *Controller) transitionTo(ctx context.Context, upgrade *upgradesv1alpha.Upgrade, phase upgradesv1alpha.UpgradePhase) error {
	p := ctrlclient.MergeFrom(upgrade.DeepCopy())
	upgrade.Status.Phase = phase
	if err := c.client.Status().Patch(ctx, upgrade, p); err != nil {
		return fmt.Errorf("failed to patch: %w", err)
	}
	return nil
}

func (c *Controller) triggerFeature(name types.FeatureName) error {
	if c.notifyFeatureController == nil {
		return fmt.Errorf("notifyFeatureController is not set, cannot trigger feature %q", name)
	}

	switch name {
	case features.Network:
		c.notifyFeatureController(true, false, false, false, false, false, false)
	case features.Gateway:
		c.notifyFeatureController(false, true, false, false, false, false, false)
	case features.Ingress:
		c.notifyFeatureController(false, false, true, false, false, false, false)
	case features.LoadBalancer:
		c.notifyFeatureController(false, false, false, true, false, false, false)
	case features.LocalStorage:
		c.notifyFeatureController(false, false, false, false, true, false, false)
	case features.MetricsServer:
		c.notifyFeatureController(false, false, false, false, false, true, false)
	case features.DNS:
		c.notifyFeatureController(false, false, false, false, false, false, true)
	default:
		return fmt.Errorf("unknown feature %q", name)
	}

	return nil
}
