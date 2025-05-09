package mountutils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/canonical/k8s/pkg/log"
)

type MockMountManager struct {
	mountsPath string
}

func (h MockMountManager) ForEachMount(ctx context.Context, callback func(ctx context.Context, device string, mountPoint string, fsType string, flags string) error) error {
	return forEachMount(ctx, h.mountsPath, callback)
}

func (h MockMountManager) Unmount(ctx context.Context, mountPoint string, flags int) error {
	log := log.FromContext(ctx)
	log.Info("Mock unmount called", "mountPoint", mountPoint, "flags", flags)

	err := h.removeLineFromMounts(mountPoint)
	if err != nil {
		return fmt.Errorf("failed to remove mount point %s from mounts: %w", mountPoint, err)
	}

	return nil
}

func (h MockMountManager) DetachLoopDevice(ctx context.Context, device string) error {
	log := log.FromContext(ctx)
	log.Info("Mock detach loop device called", "device", device)

	err := h.removeLineFromMounts(device)
	if err != nil {
		return fmt.Errorf("failed to remove loop device %s from mounts: %w", device, err)
	}

	// Simulate successful detachment without actual operation
	return nil
}

func NewMockMountHelper(mountsPath string) MountManager {
	return &MockMountManager{
		mountsPath: mountsPath,
	}
}

func (h MockMountManager) removeLineFromMounts(removed string) error {
	if h.mountsPath == "" {
		// If mountsPath is not set, we ignore the operation
		return nil
	}
	data, err := os.ReadFile(h.mountsPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, removed) {
			newLines = append(newLines, line)
		}
	}
	err = os.WriteFile(h.mountsPath, []byte(strings.Join(newLines, "\n")), 0o644)
	if err != nil {
		return err
	}

	return nil
}
