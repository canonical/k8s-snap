package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/log"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconcile implements the Reconciler interface and wraps the reconcile method.
func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	res, err := c.reconcile(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile: %w", err)
	}

	bareResult := res == ctrl.Result{}
	if bareResult {
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
	}

	return res, nil
}

// reconcile is the main reconciliation loop for the upgrade controller.
func (c *Controller) reconcile(ctx context.Context) (ctrl.Result, error) {
	log := c.logger.WithValues("step", "reconcile")

	// TODO(Hue): (KU-3215) Use mgr.Client when Upgrade CRD is created with kubebuilder.
	k8sClient, err := c.snap.KubernetesClient("")
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

	c.logger.WithValues("upgrade", upgrade.Name, "phase", upgrade.Status.Phase).Info("Reconciling upgrade.")

	switch {
	case upgrade.Status.Phase == kubernetes.UpgradePhaseNodeUpgrade:
		return c.reconcileNodeUpgrade(ctx, k8sClient, upgrade)
	case upgrade.Status.Phase == kubernetes.UpgradePhaseFeatureUpgrade:
		return c.reconcileFeatureUpgrade(ctx, k8sClient, upgrade)
	default:
		// NOTE(Hue): This should never happen, but even then we don't want to return an error.
		log.Info("Unknown upgrade phase", "phase", upgrade.Status.Phase)
		return ctrl.Result{}, nil
	}
}

// reconcileNodeUpgrade checks if all nodes have been upgraded.
// If so, it transitions to the feature upgrade phase and notifies the feature controller.
func (c *Controller) reconcileNodeUpgrade(ctx context.Context, client *kubernetes.Client, upgrade *kubernetes.Upgrade) (ctrl.Result, error) {
	log := c.logger.WithValues("upgrade", upgrade.Name, "step", "node-upgrade")

	allNodesUpgraded, err := c.allNodesUpgraded(ctx, upgrade.Status.UpgradedNodes)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to check if all nodes have been upgraded: %w", err)
	} else if !allNodesUpgraded {
		return ctrl.Result{}, nil
	}

	log.Info("All nodes have been upgraded.")

	// This will trigger another reconciliation for the upgrade object.
	if err := client.PatchUpgradeStatus(ctx, upgrade.Name, kubernetes.UpgradeStatus{Phase: kubernetes.UpgradePhaseFeatureUpgrade}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set upgrade phase: %w", err)
	}

	log.Info("Transitioned to feature-upgrade phase.")

	return ctrl.Result{}, nil
}

// reconcileFeatureUpgrade triggers feature controllers to reconcile
// and waits for them to finish.
func (c *Controller) reconcileFeatureUpgrade(ctx context.Context, client *kubernetes.Client, upgrade *kubernetes.Upgrade) (ctrl.Result, error) {
	log := c.logger.WithValues("upgrade", upgrade.Name, "step", "feature-upgrade")
	log.Info("Triggering feature controllers")

	select {
	case <-c.featureControllerReadyCh:
	case <-time.After(c.featureControllerReadyTimeout):
		return ctrl.Result{}, fmt.Errorf("timed out waiting for feature controllers to be ready")
	}

	log.Info("Waiting for feature controllers to reconcile.")

	for name, ch := range c.featureToReconciledCh {
		timeout := time.After(c.featureControllerReconcileTimeout)

		switch name {
		case string(features.Network):
			c.notifyNetworkFeature()
		case string(features.Gateway):
			c.notifyGatewayFeature()
		case string(features.Ingress):
			c.notifyIngressFeature()
		case string(features.LoadBalancer):
			c.notifyLoadBalancerFeature()
		case string(features.LocalStorage):
			c.notifyLocalStorageFeature()
		case string(features.MetricsServer):
			c.notifyMetricsServerFeature()
		case string(features.DNS):
			c.notifyDNSFeature()
		default:
			return ctrl.Result{}, fmt.Errorf("trying to reconcile unknown feature %q", name)
		}

		select {
		case <-ctx.Done():
			return ctrl.Result{}, fmt.Errorf("context done while waiting for features to get reconciled: %w", ctx.Err())
		case <-timeout:
			// TODO(Hue): (KU-3227) Do something about failed feature reconciliations.
			return ctrl.Result{}, fmt.Errorf("timed out waiting for feature %q to get reconciled", name)
		case <-ch:
			log.Info(fmt.Sprintf("feature %q have reconciled.", name))
		}
	}

	log.Info("All feature have reconciled.")

	if err := client.PatchUpgradeStatus(ctx, upgrade.Name, kubernetes.UpgradeStatus{Phase: kubernetes.UpgradePhaseCompleted}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set upgrade phase after successful feature upgrade: %w", err)
	}

	return ctrl.Result{}, nil
}

// allNodesUpgraded checks if all nodes in the cluster have been upgraded.
func (c *Controller) allNodesUpgraded(ctx context.Context, upgradedNodes []string) (bool, error) {
	log := log.FromContext(ctx)

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
