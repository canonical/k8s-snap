package app

import (
	"context"
	"fmt"
	"time"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/log"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/canonical/microcluster/v2/state"
)

// postRefreshHook is executed after the node is ready after a `snap refresh` operation
// See nodeReadyHook for details on when a node is considered ready.
// Note that the postRefreshHook is NOT executed after a `snap install` operation which is
// different to the underlying snap hook.
func (a *App) postRefreshHook(ctx context.Context, s state.State) error {
	log := log.FromContext(ctx).WithValues("hook", "post-refresh")
	log.Info("Running post-refresh hook")

	isWorker, err := snaputil.IsWorker(a.snap)
	if err != nil {
		return fmt.Errorf("failed to check if node is a worker: %w", err)
	}

	if isWorker {
		log.Info("Node is a worker, skipping post-refresh hook")
		return nil
	}

	config, err := databaseutil.GetClusterConfig(ctx, s)
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %w", err)
	}

	log.Info("Re-enable snapd/k8sd config sync and reconcile.")
	if err := snapdconfig.SetSnapdFromK8sd(ctx, config.ToUserFacing(), a.snap); err != nil {
		return fmt.Errorf("failed to set snapd configuration from k8sd: %w", err)
	}

	status, err := a.MicroCluster().Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get boostrap status: %w", err)
	}

	if _, ok := config.Annotations.Get(apiv1_annotations.AnnotationDisableSeparateFeatureUpgrades); !ok {
		// We don't want to run the upgrade if the cluster is not ready.
		// The post-refresh hook is run after snap refresh AND install, so we need to make sure the cluster is ready.
		if status.Ready {
			log.Info("Cluster is ready, running post-upgrade.")
			if err := a.performPostUpgrade(ctx, s); err != nil {
				return fmt.Errorf("failed to perform post-upgrade: %w", err)
			}
		} else {
			log.Info("Node is not yet bootstrapped (was freshly installed), skipping upgrade steps.")
		}
	} else {
		log.Info("Skipping post-upgrade steps as per user annotation.")
	}

	return nil
}

// performPostUpgrade performs the post-upgrade steps.
// It marks the node as upgraded and checks if all nodes have been upgraded.
// If all nodes have been upgraded, it triggers the feature upgrade.
func (a *App) performPostUpgrade(ctx context.Context, s state.State) error {
	log := log.FromContext(ctx).WithValues("step", "post-upgrade")
	k8sClient, err := a.snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	upgrade, err := k8sClient.GetInProgressUpgrade(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for in-progress upgrade: %w", err)
	}

	if upgrade == nil {
		log.Info("No upgrade is in progress - creating a new one.")
		rev, err := a.snap.Revision(ctx)
		if err != nil {
			return fmt.Errorf("failed to get revision: %w", err)
		}
		// TODO(ben): Add more metadata to the upgrade.
		// e.g. initial revision, target revision, name of the node that started the upgrade, etc.
		newUpgrade := kubernetes.NewUpgrade(fmt.Sprintf("cluster-upgrade-to-rev-%s", rev))
		upgrade = &newUpgrade
		if err := k8sClient.CreateUpgrade(ctx, *upgrade); err != nil {
			return fmt.Errorf("failed to create upgrade: %w", err)
		}

	} else {
		log.Info("Upgrade in progress.", "upgrade", upgrade.Metadata.Name, "phase", upgrade.Status.Phase)
	}

	log.Info("Marking node as upgraded.", "node", s.Name())

	upgradedNodes := upgrade.Status.UpgradedNodes
	upgradedNodes = append(upgradedNodes, s.Name())

	if err := k8sClient.PatchUpgradeStatus(ctx, upgrade.Metadata.Name, kubernetes.UpgradeStatus{UpgradedNodes: upgradedNodes}); err != nil {
		return fmt.Errorf("failed to mark node as upgraded: %w", err)
	}

	clusterUpgradeDone, err := allNodesUpgraded(ctx, s, upgradedNodes)
	if err != nil {
		return fmt.Errorf("failed to check if all nodes have been upgraded: %w", err)
	}

	if clusterUpgradeDone {
		log.Info("All nodes have been upgraded.")
		if err := k8sClient.PatchUpgradeStatus(ctx, upgrade.Metadata.Name, kubernetes.UpgradeStatus{Phase: kubernetes.UpgradePhaseFeatureUpgrade}); err != nil {
			return fmt.Errorf("failed to set upgrade phase: %w", err)
		}

		log.Info("Triggering feature upgrades in background")

		// TODO(ben): This is a bit ugly. We cannot wait here for the feature controllers to finish because
		// the controllers are blocked until the node is marked as ready. For this phase, we can just run a separate
		// goroutine to trigger the feature controllers and wait for them to finish.
		// Once we have a separate feature upgrade controller, this should be much cleaner.
		go func() {
			log.Info("Triggering feature controllers in separate go routine.")
			// TODO: Do we need to handle dependencies between features?
			// If yes, a new features/interface/upgrades should be created that takes all trigger and reconciled channels.
			// The custom dependencies could then be handled in the flavor feature implementation.

			// Trigger all feature controllers
			// This intentionally blocks until the feature controllers are available.
			<-a.featureController.ReadyCh()
			a.NotifyFeatureController(true, true, true, true, true, true, true)

			log.Info("Waiting for feature controllers to reconcile.")
			pending := map[string]<-chan struct{}{
				"Network":       a.featureController.ReconciledNetworkCh(),
				"Gateway":       a.featureController.ReconciledGatewayCh(),
				"Ingress":       a.featureController.ReconciledIngressCh(),
				"DNS":           a.featureController.ReconciledDNSCh(),
				"LoadBalancer":  a.featureController.ReconciledLoadBalancerCh(),
				"LocalStorage":  a.featureController.ReconciledLocalStorageCh(),
				"MetricsServer": a.featureController.ReconciledMetricsServerCh(),
			}

			for len(pending) > 0 {
				select {
				case <-ctx.Done():
					log.Error(ctx.Err(), "Context canceled while waiting for feature controllers to reconcile.")
					return

				case <-time.After(100 * time.Millisecond):
					for name, ch := range pending {
						select {
						case <-ch:
							log.Info(fmt.Sprintf("%s feature controller reconciled.", name))
							delete(pending, name)
						default:
						}
					}
				}
			}

			log.Info("All feature have reconciled.")

			if err := k8sClient.PatchUpgradeStatus(ctx, upgrade.Metadata.Name, kubernetes.UpgradeStatus{Phase: kubernetes.UpgradePhaseCompleted}); err != nil {
				log.Error(err, "failed to set upgrade phase after successful feature upgrade")
				return
			}
		}()
	}
	return nil
}

// allNodesUpgraded checks if all nodes in the cluster have been upgraded.
func allNodesUpgraded(ctx context.Context, s state.State, upgradedNodes []string) (bool, error) {
	log := log.FromContext(ctx)

	// Check if all nodes have been upgraded
	c, err := s.Leader()
	if err != nil {
		return false, fmt.Errorf("failed to get leader client: %w", err)
	}

	clusterMembers, err := c.GetClusterMembers(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get cluster members: %w", err)
	}

	if len(clusterMembers) != len(upgradedNodes) {
		log.Info("Not all nodes have been upgraded.", "clusterMembers", len(clusterMembers), "upgradedNodes", len(upgradedNodes))
		return false, nil
	}

	clusterMembersMap := make(map[string]struct{})
	for _, member := range clusterMembers {
		clusterMembersMap[member.Name] = struct{}{}
	}

	for _, node := range upgradedNodes {
		if _, ok := clusterMembersMap[node]; !ok {
			log.Info("Node has been upgraded but is not part of the cluster.", "node", node)
			return false, nil
		}
	}
	return true, nil
}
