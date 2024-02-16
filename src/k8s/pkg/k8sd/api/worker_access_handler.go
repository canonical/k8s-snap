package api

import (
	"fmt"
	"net/http"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func RestrictWorkers(s *state.State, r *http.Request) response.Response {
	snap := snap.SnapFromContext(s.Context)

	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}

	if isWorker {
		return response.Forbidden(fmt.Errorf("this action is restricted on workers"))
	}

	return response.EmptySyncResponse
}
