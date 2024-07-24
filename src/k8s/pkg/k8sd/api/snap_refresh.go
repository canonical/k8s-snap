package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) postSnapRefresh(s *state.State, r *http.Request) response.Response {
	req := apiv1.SnapRefreshRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	refreshOpts, err := types.RefreshOptsFromAPI(req)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid refresh options: %w", err))
	}
	log := log.FromContext(s.Context).WithValues("to", refreshOpts)

	readyCh := make(chan error)
	go func() {
		// block until we have flushed the response
		if err := <-readyCh; err != nil {
			log.Error(err, "Cancel refresh")
			return
		}

		log.Info("Refreshing snap")
		if err := e.provider.Snap().Refresh(s.Context, refreshOpts); err != nil {
			log.Error(err, "Failed to refresh snap")
		}
	}()

	return response.ManualResponse(func(w http.ResponseWriter) (rerr error) {
		defer func() {
			readyCh <- rerr
			close(readyCh)
		}()

		err := response.EmptySyncResponse.Render(w)
		if err != nil {
			return err
		}

		// Send the response before replacing the LXD daemon process.
		f, ok := w.(http.Flusher)
		if !ok {
			return fmt.Errorf("ResponseWriter is not type http.Flusher")
		}

		f.Flush()
		return nil
	})
}
