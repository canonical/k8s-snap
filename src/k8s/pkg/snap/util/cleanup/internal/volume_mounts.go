package internal

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	mountutils "github.com/canonical/k8s/pkg/utils/mount"
	"golang.org/x/sys/unix"
)

func containerdMountPrefixes(ctx context.Context, s snap.Snap) []string {
	paths := []string{}

	// Get the containerd root directory
	// e.g. /var/lib/containerd
	containerdRootDir := s.ContainerdRootDir()
	if snaputil.IsContainerdPathManaged(ctx, s, containerdRootDir) {
		paths = append(paths, containerdRootDir)
	}

	// e.g. /run/containerd/io.containerd.
	containerdSocketDir := s.ContainerdSocketDir()
	if snaputil.IsContainerdPathManaged(ctx, s, containerdSocketDir) {
		paths = append(paths, filepath.Join(containerdSocketDir, "io.containerd."))
	}

	return paths
}

func RemoveVolumeMountsGracefully(ctx context.Context, s snap.Snap, mountHelper mountutils.MountManager) {
	log := log.FromContext(ctx)

	prefixes := []string{"/var/lib/kubelet/pods"}
	containerdPaths := containerdMountPrefixes(ctx, s)
	prefixes = append(prefixes, containerdPaths...)
	log.Info("Removing volume mounts", "prefixes", prefixes)

	err := mountHelper.ForEachMount(ctx, func(ctx context.Context, device string, mountPoint string, fsType string, flags string) error {
		for _, prefix := range prefixes {
			if strings.HasPrefix(mountPoint, prefix) && !strings.Contains(fsType, "nfs") {
				// unmount Pod NFS volumes only forcefully, as unmounting them normally may hang otherwise.
				// unmount remaining Pod volumes gracefully.
				return mountHelper.Unmount(ctx, mountPoint, unix.MNT_DETACH)
			}
		}

		return nil
	})
	if err != nil {
		log.Error(err, "failed to iterate mounts for removing volume mounts")
	}
}

func RemoveVolumeMountsForce(ctx context.Context, s snap.Snap, mountHelper mountutils.MountManager) {
	log := log.FromContext(ctx)

	prefixes := []string{"/var/lib/kubelet/pods", "/var/lib/kubelet/plugins"}
	containerdPaths := containerdMountPrefixes(ctx, s)
	prefixes = append(prefixes, containerdPaths...)
	log.Info("Removing volume mounts", "prefixes", prefixes)

	err := mountHelper.ForEachMount(ctx, func(ctx context.Context, device string, mountPoint string, fsType string, flags string) error {
		for _, prefix := range prefixes {
			if strings.HasPrefix(mountPoint, prefix) {
				// unmount lingering Pod volumes by force, to prevent potential volume leaks.
				return mountHelper.Unmount(ctx, mountPoint, unix.MNT_FORCE)
			}
		}

		return nil
	})
	if err != nil {
		log.Error(err, "failed to iterate mounts for forcefully removing volume mounts")
	}
}
