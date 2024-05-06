package api

import (
	"errors"
	"net/http"
	"os"

	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) deleteCluster(s *state.State, r *http.Request) response.Response {
	var errs []error

	snap := e.provider.Snap()

	// Stop k8s-dqlite service and remove directory
	if err := snaputil.StopK8sDqliteServices(s.Context, snap); err != nil {
		errs = append(errs, err)
	}
	if err := os.RemoveAll(snap.K8sDqliteStateDir()); err != nil {
		errs = append(errs, err)
	}

	// Stop control plane services and remove directories
	if err := snaputil.StopControlPlaneServices(s.Context, snap); err != nil {
		errs = append(errs, err)
	}
	// clear kubernetes config dir and subdirs: PKI, ETCD PKI with certificates
	if err := os.RemoveAll(snap.KubernetesConfigDir()); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return response.InternalError(errors.Join(errs...))
	}

	return response.SyncResponse(true, nil)
}
