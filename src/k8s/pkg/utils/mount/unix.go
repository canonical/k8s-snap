package mountutils

import (
	"context"
	"fmt"
	"os/exec"

	"golang.org/x/sys/unix"
)

type UnixMountManager struct{}

func (h UnixMountManager) ForEachMount(ctx context.Context, callback func(ctx context.Context, device string, mountPoint string, fsType string, flags string) error) error {
	return forEachMount(ctx, "/proc/mounts", callback)
}

func (h UnixMountManager) Unmount(ctx context.Context, mountPoint string, flags int) error {
	if err := unix.Unmount(mountPoint, flags); err != nil {
		return fmt.Errorf("failed to unmount %s: %w", mountPoint, err)
	}
	return nil
}

func (h UnixMountManager) DetachLoopDevice(ctx context.Context, device string) error {
	cmd := exec.CommandContext(ctx, "losetup", "-d", device)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to detach loop device using losetup %s: %w", device, err)
	}
	return nil
}

func NewUnixMountHelper() MountManager {
	return &UnixMountManager{}
}
