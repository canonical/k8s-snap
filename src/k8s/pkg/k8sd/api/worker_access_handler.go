package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
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

// TokenAuthentication access handler checks if the worker is allowed to access this endpoint with the provided token.
// TokenAuthentication expects the "worker-name" and "worker-secret" fields to be set in the request header.
func TokenAuthentication(s *state.State, r *http.Request) response.Response {
	name := r.Header.Get("worker-name")
	if name == "" {
		return response.Unauthorized(fmt.Errorf("missing 'worker-name'"))
	}
	hostname, err := utils.CleanHostname(name)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", hostname, err))
	}

	token := r.Header.Get("worker-secret")
	if token == "" {
		return response.Unauthorized(fmt.Errorf("invalid token"))
	}

	var tokenIsValid bool
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		tokenIsValid, err = database.CheckWorkerNodeToken(ctx, tx, hostname, token)
		if err != nil {
			return fmt.Errorf("failed to check worker node token: %w", err)
		}
		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("check token database transaction failed: %w", err))
	}
	if !tokenIsValid {
		return response.Unauthorized(fmt.Errorf("invalid token"))
	}

	return response.EmptySyncResponse
}
