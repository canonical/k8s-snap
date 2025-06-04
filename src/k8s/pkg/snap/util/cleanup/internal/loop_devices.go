package internal

import (
	"context"
	"strings"

	"github.com/canonical/k8s/pkg/log"
	mountutils "github.com/canonical/k8s/pkg/utils/mount"
)

func RemoveLoopDevices(ctx context.Context, mountHelper mountutils.MountManager) {
	log := log.FromContext(ctx)

	err := mountHelper.ForEachMount(ctx, func(ctx context.Context, device string, mountPoint string, fsType string, flags string) error {
		if strings.HasPrefix(mountPoint, "/var/lib/kubelet/pods") && strings.HasPrefix(device, "/dev/loop") {
			// Gather loop devices for detachment
			return mountHelper.DetachLoopDevice(ctx, device)
		}

		return nil
	})
	if err != nil {
		log.Error(err, "failed to iterate mounts for loop devices")
	}
}
