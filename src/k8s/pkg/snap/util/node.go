package snaputil

import (
	"fmt"
	"os"
	"path"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

func IsWorker(snap snap.Snap) (bool, error) {
	return utils.FileExists(path.Join(snap.LockFilesDir(), "worker"))
}

func MarkAsWorkerNode(snap snap.Snap, mark bool) error {
	fname := path.Join(snap.LockFilesDir(), "worker")

	if mark {
		lock, err := os.Create(fname)
		if err != nil {
			return fmt.Errorf("failed to mark node as worker: %w", err)
		}

		defer func() {
			cErr := lock.Close()
			if cErr != nil && err == nil {
				err = cErr
			}
		}()

		if err := os.Chown(fname, snap.UID(), snap.GID()); err != nil {
			return fmt.Errorf("failed to chown %s: %w", fname, err)
		}
		if err := os.Chmod(fname, 0600); err != nil {
			return fmt.Errorf("failed to chmod %s: %w", fname, err)
		}
	} else {
		if err := os.Remove(fname); err != nil {
			return fmt.Errorf("failed to unmark node as worker: %w", err)
		}
	}

	return nil
}
