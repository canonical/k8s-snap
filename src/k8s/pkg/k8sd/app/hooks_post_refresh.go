package app

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations"
	upgradesv1alpha "github.com/canonical/k8s/pkg/k8sd/crds/upgrades/v1alpha"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/log"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	upgradepkg "github.com/canonical/k8s/pkg/upgrade"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/canonical/k8s/pkg/version"
	"github.com/canonical/microcluster/v2/state"
)

// postRefreshHook is executed after the node is ready after a `snap refresh` operation
// See nodeReadyHook for details on when a node is considered ready.
// Note that the postRefreshHook is NOT executed after a `snap install` operation which is
// different to the underlying snap hook.
func (a *App) postRefreshHook(ctx context.Context, s state.State) error {
	log := log.FromContext(ctx).WithValues("hook", "post-refresh")
	log.Info("Running post-refresh hook")

	if err := a.updateNodeIPAddresses(ctx, s); err != nil {
		return fmt.Errorf("failed to update node IP addresses: %w", err)
	}

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
		log.Info("Post-upgrade steps skipped due to user annotation override.")
	}

	return nil
}

// performPostUpgrade adds the node name to the list of upgradedNodes in the upgrade custom resource.
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

		k8sVersion, err := a.snap.NodeKubernetesVersion(ctx)
		if err != nil {
			return fmt.Errorf("failed to get kubernetes version: %w", err)
		}

		// TODO(ben): Add more metadata to the upgrade.
		// e.g. initial revision, target revision, name of the node that started the upgrade, etc.
		versionData := version.Info{Revision: rev, KubernetesVersion: k8sVersion}
		upgrade = upgradesv1alpha.NewUpgrade(upgradepkg.GetName(versionData))
		if err := k8sClient.Create(ctx, upgrade); err != nil {
			return fmt.Errorf("failed to create upgrade: %w", err)
		}
		log.Info("Created new upgrade CR.", "upgrade", *upgrade)
	} else {
		log.Info("Upgrade in progress.", "upgrade", upgrade.Name, "phase", upgrade.Status.Phase)
	}

	log.Info("Marking node as upgraded.", "node", s.Name())

	upgradedNodes := upgrade.Status.UpgradedNodes
	if !slices.Contains(upgradedNodes, s.Name()) {
		upgradedNodes = append(upgradedNodes, s.Name())
	}

	status := upgradesv1alpha.UpgradeStatus{
		UpgradedNodes: upgradedNodes,
		Phase:         upgradesv1alpha.UpgradePhaseNodeUpgrade,
		Strategy:      upgradesv1alpha.UpgradeStrategyInPlace,
	}

	if err := k8sClient.PatchUpgradeStatus(ctx, upgrade, status); err != nil {
		return fmt.Errorf("failed to mark node as upgraded: %w", err)
	}

	log.Info("Marked node as upgraded.", "status", upgrade.Status)

	return nil
}

func (a *App) updateNodeIPAddresses(ctx context.Context, s state.State) error {
	snap := a.Snap()
	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return fmt.Errorf("failed to parse node IP address %s", s.Address().Hostname())
	}

	// nodeIPs will be passed to kubelet as the --node-ip parameter, allowing it to have multiple node IPs,
	// including IPv4 and IPv6 addresses for dualstacks.
	nodeIPs, err := utils.GetIPv46Addresses(nodeIP)
	if err != nil {
		return fmt.Errorf("failed to get local node IPs for kubelet: %w", err)
	}

	ips := []string{}
	for _, nodeIP := range nodeIPs {
		if !nodeIP.IsLoopback() {
			ips = append(ips, nodeIP.String())
		}
	}
	if len(ips) > 0 {
		args := map[string]string{"--node-ip": strings.Join(ips, ",")}
		mustRestartKubelet, err := snaputil.UpdateServiceArguments(snap, "kubelet", args, nil)
		if err != nil {
			return fmt.Errorf("failed to render arguments file: %w", err)
		}

		if mustRestartKubelet {
			// This may fail if other controllers try to restart the services at the same time, hence the retry.
			if err := control.RetryFor(ctx, 5, 5*time.Second, func() error {
				if err := snap.RestartServices(ctx, []string{"kubelet"}); err != nil {
					return fmt.Errorf("failed to restart kubelet to apply node ip addresses: %w", err)
				}
				return nil
			}); err != nil {
				return fmt.Errorf("failed after retry: %w", err)
			}
		}
	}

	return nil
}
