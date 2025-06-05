package internal

import (
	"context"
	"os"
	"strconv"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils/shims"
)

func RemoveContainers(ctx context.Context) {
	log := log.FromContext(ctx)
	pids, err := shims.RunningContainerdShimPIDs(ctx)
	if err != nil {
		log.Error(err, "failed to get containerd shim PIDs")
		return
	}

	for _, pid := range pids {
		intPid, err := strconv.Atoi(pid)
		if err != nil {
			log.Error(err, "failed to convert PID to integer", "pid", pid)
			continue
		}

		process, err := os.FindProcess(intPid)
		if err != nil {
			log.Error(err, "failed to find containerd shim PID", "pid", intPid)
			continue
		}

		if err := process.Kill(); err != nil {
			log.Error(err, "failed to kill containerd shim PID", "pid", intPid)
			continue
		}
	}
}
