package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/log"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconcile is the main reconciliation loop for the upgrade controller.
func (r *upgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Logger.WithValues("upgrade", req.Name)

	// TODO(Hue): (KU-3215) Use mgr.Client when Upgrade CRD is created with kubebuilder.
	k8sClient, err := r.snap.KubernetesClient("")
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	upgrade, err := k8sClient.GetInProgressUpgrade(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to check for in-progress upgrade: %w", err)
	}

	if upgrade == nil {
		log.Info("No in-progress upgrade found, ignoring")
		return ctrl.Result{}, nil
	}

	switch {
	case upgrade.Status.Phase == kubernetes.UpgradePhaseNodeUpgrade:
		return r.reconcileNodeUpgrade(ctx, k8sClient, upgrade)
	case upgrade.Status.Phase == kubernetes.UpgradePhaseFeatureUpgrade:
		return r.reconcileFeatureUpgrade(ctx, k8sClient, upgrade)
	default:
		// NOTE(Hue): This should never happen, but even then we don't want to return an error.
		log.Info("Unknown upgrade phase", "phase", upgrade.Status.Phase)
		return ctrl.Result{}, nil
	}
}

// reconcileNodeUpgrade checks if all nodes have been upgraded.
// If so, it transitions to the feature upgrade phase and notifies the feature controller.
func (r *upgradeReconciler) reconcileNodeUpgrade(ctx context.Context, c *kubernetes.Client, upgrade *kubernetes.Upgrade) (ctrl.Result, error) {
	log := r.Logger.WithValues("upgrade", upgrade.Name, "step", "node-upgrade")

	allNodesUpgraded, err := r.allNodesUpgraded(ctx, upgrade.Status.UpgradedNodes)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to check if all nodes have been upgraded: %w", err)
	} else if !allNodesUpgraded {
		// NOTE(Hue): In case a node left the cluster during an upgrade
		// and the upgrade is finished sooner than expected. We need this requeue
		// since nothing will change the upgrade CR again.
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	log.Info("All nodes have been upgraded.")

	// This will trigger another reconciliation for the upgrade object.
	if err := c.PatchUpgradeStatus(ctx, upgrade.Name, kubernetes.UpgradeStatus{Phase: kubernetes.UpgradePhaseFeatureUpgrade}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set upgrade phase: %w", err)
	}

	log.Info("Transitioned to feature-upgrade phase.")

	return ctrl.Result{}, nil
}

// reconcileFeatureUpgrade triggers feature controllers to reconcile
// and waits for them to finish.
func (r *upgradeReconciler) reconcileFeatureUpgrade(ctx context.Context, c *kubernetes.Client, upgrade *kubernetes.Upgrade) (ctrl.Result, error) {
	log := r.Logger.WithValues("upgrade", upgrade.Name, "step", "feature-upgrade")
	log.Info("Triggering feature controllers")

	select {
	case <-r.featureControllerReadyCh:
	case <-time.After(r.featureControllerReadyTimeout):
		return ctrl.Result{}, fmt.Errorf("timed out waiting for feature controllers to be ready")
	}

	r.notifyFeatureController()

	log.Info("Waiting for feature controllers to reconcile.")

	ctx, cancel := context.WithTimeout(ctx, r.featureControllerReconcileTimeout)
	defer cancel()

	for name, ch := range r.featureToReconciledCh {
		select {
		case <-ctx.Done():
			return ctrl.Result{}, fmt.Errorf("context done while waiting for feature controllers to reconcile: %w", ctx.Err())
		case <-ch:
			log.Info(fmt.Sprintf("%s feature controller reconciled.", name))
		}
	}

	log.Info("All feature have reconciled.")

	if err := c.PatchUpgradeStatus(ctx, upgrade.Name, kubernetes.UpgradeStatus{Phase: kubernetes.UpgradePhaseCompleted}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set upgrade phase after successful feature upgrade: %w", err)
	}

	return ctrl.Result{}, nil
}

// allNodesUpgraded checks if all nodes in the cluster have been upgraded.
func (r *upgradeReconciler) allNodesUpgraded(ctx context.Context, upgradedNodes []string) (bool, error) {
	log := log.FromContext(ctx)

	c, err := r.getState().Leader()
	if err != nil {
		return false, fmt.Errorf("failed to get leader client: %w", err)
	}

	clusterMembers, err := c.GetClusterMembers(ctx)
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
