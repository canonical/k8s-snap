package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func ValidateCAPIAuthTokenAccessHandler(tokenHeaderName string) func(s *state.State, r *http.Request) response.Response {
	return func(s *state.State, r *http.Request) response.Response {
		token := r.Header.Get(tokenHeaderName)
		if token == "" {
			return response.Unauthorized(fmt.Errorf("missing header %q", tokenHeaderName))
		}

		var tokenIsValid bool
		if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
			var err error
			tokenIsValid, err = database.ValidateClusterAPIToken(ctx, tx, token)
			if err != nil {
				return fmt.Errorf("failed to check CAPI auth token: %w", err)
			}
			return nil
		}); err != nil {
			return response.InternalError(fmt.Errorf("check CAPI auth token database transaction failed: %w", err))
		}
		if !tokenIsValid {
			return response.Unauthorized(fmt.Errorf("invalid token"))
		}

		return response.EmptySyncResponse
	}
}
