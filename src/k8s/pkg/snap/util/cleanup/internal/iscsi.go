package internal

import (
	"context"
	"errors"
	"os/exec"

	"github.com/canonical/k8s/pkg/log"
)

// iscsiadmNoObjectsFound is the exit code returned by iscsiadm when no iSCSI sessions exist.
// Defined as ISCSI_ERR_NO_OBJS_FOUND in open-iscsi.
// https://github.com/open-iscsi/open-iscsi/blob/2.1.11/include/iscsi_err.h#L50
const iscsiadmNoObjectsFound = 21

// LogoutISCSISessions logs out all active iSCSI sessions, allowing volume unmounts to proceed
// without blocking on the kernel's iSCSI session recovery timeout (default: 120s).
// This is required when iSCSI-backed volumes (e.g. Longhorn) are present and the iSCSI target
// becomes unreachable after services are stopped.
func LogoutISCSISessions(ctx context.Context) {
	log := log.FromContext(ctx)

	if _, err := exec.LookPath("iscsiadm"); err != nil {
		log.Info("iscsiadm not found, skipping iSCSI session logout")
		return
	}

	out, err := exec.CommandContext(ctx, "iscsiadm", "-m", "session", "-u").CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == iscsiadmNoObjectsFound {
			return
		}
		log.Error(err, "failed to logout iSCSI sessions", "output", string(out))
	}
}
