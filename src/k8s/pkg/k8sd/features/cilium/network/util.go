package network

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils"
)

func (r reconciler) verifyMountPropagation(ctx context.Context) error {
	snap := r.Snap()

	pt, err := GetMountPropagationType("/sys")
	if err != nil {
		return fmt.Errorf("failed to get mount propagation type for /sys: %w", err)
	}

	if pt == utils.MountPropagationPrivate {
		onLXD, err := snap.OnLXD(ctx)
		if err != nil {
			logger := log.FromContext(ctx)
			logger.Error(err, "Failed to check if running on LXD")
		}
		if onLXD {
			return fmt.Errorf("/sys is not a shared mount on the LXD container, this might be resolved by updating LXD on the host to version 5.0.2 or newer")
		}

		return fmt.Errorf("/sys is not a shared mount")
	}

	return nil
}
