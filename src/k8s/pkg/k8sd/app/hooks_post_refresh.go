package app

import (
	"context"
	"fmt"

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

	k8sClient, err := a.snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	upgrade, err := k8sClient.GetInProgressUpgrade(ctx)
	if err != nil {
		return fmt.Errorf("failed to get in progress upgrade: %w", err)
	}

	if upgrade != nil {
		log.Info("Upgrade in progress.", "upgrade", upgrade.Metadata.Name, "phase", upgrade.Status.Phase)
		log.Info("Marking node as upgraded.", "node", s.Name())

		if err := k8sClient.MarkNodeUpgradeDone(ctx, s.Name()); err != nil {
			return fmt.Errorf("failed to mark node as upgraded: %w", err)
		}

		clusterUpgradeDone, err := allNodesUpgraded(ctx, s, k8sClient)
		if err != nil {
			return fmt.Errorf("failed to check if all nodes have been upgraded: %w", err)
		}

		if clusterUpgradeDone {
			log.Info("All nodes have been upgraded.")
			if err := k8sClient.SetUpgradePhase(ctx, "FeatureUpgrade"); err != nil {
				return fmt.Errorf("failed to set upgrade phase: %w", err)
			}

			// TODO: Trigger feature upgrade and unlock feature controllers afterwards.
		}
	} else {
		log.Info("No upgrade is in progress.")
	}

	return nil
}

// allNodesUpgraded checks if all nodes in the cluster have been upgraded.
func allNodesUpgraded(ctx context.Context, s state.State, k8sClient *kubernetes.Client) (bool, error) {
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

	upgradedNodes, err := k8sClient.GetUpgradedNodes(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get upgraded nodes: %w", err)
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
