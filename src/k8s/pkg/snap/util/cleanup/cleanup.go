package cleanup

import (
	"context"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/snap/util/cleanup/internal"
	mountutils "github.com/canonical/k8s/pkg/utils/mount"
	netnsutils "github.com/canonical/k8s/pkg/utils/netns"
)

// [DANGER] Cleanup containers and runtime state. Note that the order of operations below is crucial.
// Cleanup is done on a best-effort basis, and errors are logged but not returned.
func TryCleanupContainers(ctx context.Context, s snap.Snap) {
	mountHelper := mountutils.NewUnixMountHelper()
	netnsHelper := netnsutils.NewUnixNetworkNSHelper()

	internal.RemoveContainers(ctx)
	internal.RemoveNetworkNamespaces(ctx, netnsHelper)
	internal.RemoveVolumeMountsGracefully(ctx, s, mountHelper)
	internal.RemoveVolumeMountsForce(ctx, s, mountHelper)
	internal.RemovePluginSockets(ctx)
	internal.RemoveLoopDevices(ctx, mountHelper)
}

// TryCleanupContainerdPaths attempts to clean up all containerd directories which were
// created by the k8s-snap based on the existence of their respective lockfiles
// located in the directory returned by `s.LockFilesDir()`.
func TryCleanupContainerdPaths(ctx context.Context, s snap.Snap) {
	internal.TryCleanupContainerdPaths(ctx, s)
}

// RemoveKubeProxyRules removes routing rules, such as iptables rules, that were created by kube-proxy.
func RemoveKubeProxyRules(ctx context.Context, s snap.Snap) {
	internal.RemoveKubeProxyRules(ctx, s)
}
