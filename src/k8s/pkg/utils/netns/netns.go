package netnsutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/log"
	mountutils "github.com/canonical/k8s/pkg/utils/mount"
	"golang.org/x/sys/unix"
)

type NetworkNamespaceManager interface {
	ForEachNetworkNamespace(ctx context.Context, callback func(ctx context.Context, namespace string) error) error
	DeleteNetworkNamespace(ctx context.Context, namespace string) error
}

func deleteNetworkNamespace(ctx context.Context, mountHelper mountutils.MountManager, netnsDir string, namespace string) error {
	nsPath := filepath.Join(netnsDir, namespace)

	if err := mountHelper.Unmount(ctx, nsPath, unix.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount network namespace %s: %w", namespace, err)
	}

	if err := os.Remove(nsPath); err != nil {
		return fmt.Errorf("failed to remove network namespace %s: %w", namespace, err)
	}

	return nil
}

func forEachNetworkNamespace(ctx context.Context, netnsDir string, callback func(ctx context.Context, namespace string) error) error {
	log := log.FromContext(ctx)

	entries, err := os.ReadDir(netnsDir)
	if err != nil {
		return fmt.Errorf("failed to list files under network namespace directory %w", err)
	}

	for _, entry := range entries {
		if err := callback(ctx, entry.Name()); err != nil {
			log.Error(err, "callback failed for network namespace entry", "namespace", entry.Name())
		}
	}

	return nil
}
