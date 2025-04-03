package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) getNodeStatus(s state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	status, err := impl.GetLocalNodeStatus(r.Context(), s, snap)
	if err != nil {
		return response.InternalError(err)
	}

	taints, err := getNodeTaints(snap)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get node taints: %w", err))
	}

	return response.SyncResponse(true, &apiv1.NodeStatusResponse{
		NodeStatus: status,
		Taints:     taints,
	})
}

// getNodeTaints retrieves the taints of the local node.
func getNodeTaints(snap snap.Snap) ([]string, error) {
	taintsStr, err := snaputil.GetServiceArgument(snap, "kubelet", "--register-with-taints")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get kubelet taints: %w", err)
	}

	return strings.Split(taintsStr, ","), nil
}
