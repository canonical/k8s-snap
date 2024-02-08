package snaputil

import (
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

func IsWorker(s snap.Snap) (bool, error) {
	return utils.FileExists(s.WorkerNodeLockFile())
}
