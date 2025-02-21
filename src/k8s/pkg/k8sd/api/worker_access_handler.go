package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/canonical/k8s/pkg/k8sd/database"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) restrictWorkers(s state.State, r *http.Request) (bool, response.Response) {
	snap := e.provider.Snap()

	isWorker, err := snaputil.IsWorker(snap)
	if err != nil {
		return false, response.InternalError(fmt.Errorf("failed to check if node is a worker: %w", err))
	}

	if isWorker {
		return false, response.Forbidden(fmt.Errorf("this action is restricted on workers"))
	}

	return true, nil
}

// ValidateWorkerInfoAccessHandler access handler checks if the worker is allowed to access this endpoint with the provided token.
func ValidateWorkerInfoAccessHandler(nodeHeaderName string, tokenHeaderName string) func(s state.State, r *http.Request) (bool, response.Response) {
	return func(s state.State, r *http.Request) (bool, response.Response) {
		name := r.Header.Get(nodeHeaderName)
		if name == "" {
			return false, response.Unauthorized(fmt.Errorf("missing header %q", nodeHeaderName))
		}
		hostname, err := utils.CleanHostname(name)
		if err != nil {
			return false, response.BadRequest(fmt.Errorf("invalid hostname %q: %w", hostname, err))
		}

		token := r.Header.Get(tokenHeaderName)
		if token == "" {
			return false, response.Unauthorized(fmt.Errorf("invalid token"))
		}

		var tokenIsValid bool
		if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
			var err error
			tokenIsValid, err = database.CheckWorkerNodeToken(ctx, tx, hostname, token)
			if err != nil {
				return fmt.Errorf("failed to check worker node token: %w", err)
			}
			return nil
		}); err != nil {
			return false, response.InternalError(fmt.Errorf("check token database transaction failed: %w", err))
		}
		if !tokenIsValid {
			return false, response.Unauthorized(fmt.Errorf("invalid token"))
		}

		return true, nil
	}
}
