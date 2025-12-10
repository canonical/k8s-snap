package upgrade

import (
	"context"
	"fmt"
	"slices"
	"time"

	upgradesv1alpha "github.com/canonical/k8s-snap-api/api/v1alpha"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	upgradepkg "github.com/canonical/k8s/pkg/upgrade"
	"github.com/canonical/k8s/pkg/version"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconcile implements the Reconciler interface and wraps the reconcile method.
// Reconcile ensures that the reconciliation is requeued.
func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	res, err := c.reconcile(ctx, req)
	if err != nil {
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
	errs := []error{}
	var upgrade upgradesv1alpha.Upgrade
	if err := c.client.Get(ctx, req.NamespacedName, &upgrade); err == nil {
		res, err := c.reconcileUpgrade(ctx, &upgrade)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to reconcile upgrade %q: %w", req.NamespacedName, err)
		}
		return res, nil
	} else if !apierrors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("failed to get upgrade %q: %w", req.NamespacedName, err))
	}

	var node corev1.Node
	if err := c.client.Get(ctx, req.NamespacedName, &node); err == nil {
		res, err := c.reconcileNode(ctx, &node)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to reconcile node %q: %w", req.NamespacedName, err)
		}
		return res, nil
	} else if !apierrors.IsNotFound(err) {
		errs = append(errs, fmt.Errorf("failed to get node %q: %w", req.NamespacedName, err))
	}

	if len(errs) == 0 {
		c.logger.Info("No upgrade or node found to reconcile", "resource", req.NamespacedName)
	}

	return ctrl.Result{}, errors.NewAggregate(errs)
}

// reconcileNode handles the reconciliation of a node resource.
func (c *Controller) reconcileNode(ctx context.Context, node *corev1.Node) (ctrl.Result, error) {
	vStr, ok := node.Annotations[version.NodeAnnotationKey]
	if !ok {
		c.logger.WithValues("node", node.Name).Info("Node does not have a version annotation, skipping reconciliation.")
		return ctrl.Result{}, nil
	}

	var versionData version.Info
	if err := versionData.Decode([]byte(vStr)); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to decode version info for node %q: %w", node.Name, err)
	}

	c.logger.WithValues("node", node.Name, "version", versionData.Revision).Info("Reconciling node version.")

	var upgrade upgradesv1alpha.Upgrade
	if err := c.client.Get(ctx, ctrlclient.ObjectKey{
		Name: upgradepkg.GetName(versionData),
	}, &upgrade); err != nil && apierrors.IsNotFound(err) {
		c.logger.WithValues("node", node.Name, "revision", versionData.Revision).Info("No upgrade found for revision, skipping reconciliation.")
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get upgrade for revision %q: %w", versionData.Revision, err)
	}

	if err := c.addToUpgradedNodes(ctx, &upgrade, node); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to add node %q to upgraded nodes: %w", node.Name, err)
	}

	c.logger.WithValues("node", node.Name, "upgrade", upgrade.Name).Info("Node reconciled successfully.")

	return ctrl.Result{}, nil
}

// reconcileUpgrade handles the reconciliation of an upgrade resource.
func (c *Controller) reconcileUpgrade(ctx context.Context, upgrade *upgradesv1alpha.Upgrade) (ctrl.Result, error) {
	c.logger.WithValues("upgrade", upgrade.Name, "phase", upgrade.Status.Phase).Info("Reconciling upgrade.")

	switch upgrade.Status.Phase {
	case upgradesv1alpha.UpgradePhaseNodeUpgrade:
		return c.reconcileStatusNodeUpgrade(ctx, upgrade)
	case upgradesv1alpha.UpgradePhaseFeatureUpgrade:
		return c.reconcileStatusFeatureUpgrade(ctx, upgrade)
	case upgradesv1alpha.UpgradePhaseCompleted:
		c.logger.WithValues("upgrade", upgrade.Name).Info("Upgrade completed.")
	case upgradesv1alpha.UpgradePhaseFailed:
		// TODO(Hue): (KU-3850) Include the failure reason (error) in the upgrade status.
		c.logger.WithValues("upgrade", upgrade.Name).Error(nil, "Upgrade has failed.")
		return ctrl.Result{}, fmt.Errorf("upgrade %q has failed", upgrade.Name)
	}

	return ctrl.Result{}, nil
}

// reconcileStatusNodeUpgrade checks if all nodes have been upgraded.
// If so, it transitions to the feature upgrade phase and notifies the feature controller.
func (c *Controller) reconcileStatusNodeUpgrade(ctx context.Context, upgrade *upgradesv1alpha.Upgrade) (ctrl.Result, error) {
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

// reconcileStatusFeatureUpgrade triggers feature controllers to reconcile
// and waits for them to finish.
func (c *Controller) reconcileStatusFeatureUpgrade(ctx context.Context, upgrade *upgradesv1alpha.Upgrade) (ctrl.Result, error) {
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

	// NOTE(Hue): Can Microcluster member name be different from the Kubernetes node name?
	// Yes, but we secretly rely on the fact that they are the same.
	// (KU-3749) This should be fixed in the future.
	nodeList := &corev1.NodeList{}
	if err := c.client.List(ctx, nodeList); err != nil {
		return false, fmt.Errorf("failed to list nodes: %w", err)
	}

	upgradedNodesMap := make(map[string]struct{})
	for _, n := range upgradedNodes {
		upgradedNodesMap[n] = struct{}{}
	}

	oldNodes := []string{}
	for _, node := range nodeList.Items {
		if _, ok := upgradedNodesMap[node.Name]; !ok {
			oldNodes = append(oldNodes, node.Name)
		}
	}

	if len(oldNodes) > 0 {
		log.Info("Some nodes are not upgraded yet", "oldNodes", oldNodes)
		return false, nil
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

// addToUpgradedNodes adds the given node to the list of upgraded nodes in the upgrade resource.
func (c *Controller) addToUpgradedNodes(ctx context.Context, upgrade *upgradesv1alpha.Upgrade, node *corev1.Node) error {
	var p ctrlclient.Patch
	if !slices.Contains(upgrade.Status.UpgradedNodes, node.Name) {
		p = ctrlclient.MergeFrom(upgrade.DeepCopy())
		upgrade.Status.UpgradedNodes = append(upgrade.Status.UpgradedNodes, node.Name)
	} else {
		c.logger.WithValues("node", node.Name, "upgrade", upgrade.Name).Info("Node already in upgraded nodes list, skipping.")
		return nil
	}
	if err := c.client.Status().Patch(ctx, upgrade, p); err != nil {
		return fmt.Errorf("failed to patch: %w", err)
	}
	return nil
}

func (c *Controller) triggerFeature(name types.FeatureName) error {
	switch name {
	case features.Network:
		c.notifyNetworkFeature()
	case features.Gateway:
		c.notifyGatewayFeature()
	case features.Ingress:
		c.notifyIngressFeature()
	case features.LoadBalancer:
		c.notifyLoadBalancerFeature()
	case features.LocalStorage:
		c.notifyLocalStorageFeature()
	case features.MetricsServer:
		c.notifyMetricsServerFeature()
	case features.DNS:
		c.notifyDNSFeature()
	default:
		return fmt.Errorf("unknown feature %q", name)
	}

	return nil
}
