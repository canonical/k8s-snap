package snaputil

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

func IsWorker(s snap.Snap) (bool, error) {
	return utils.FileExists(s.WorkerNodeLockFile())
}

func MarkAsWorker(s snap.Snap) error {
	lock, err := os.Create(s.WorkerNodeLockFile())
	defer lock.Close()
	if err != nil {
		return fmt.Errorf("failed to mark node as worker: %w", err)
	}
	return nil
}

func RemoveWorkerLock(s snap.Snap) error {
	err := os.Remove(s.WorkerNodeLockFile())
	if err != nil {
		return fmt.Errorf("failed to remove worker lock: %w", err)
	}
	return nil
}
