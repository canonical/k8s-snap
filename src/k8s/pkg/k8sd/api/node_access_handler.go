package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) ValidateNodeTokenAccessHandler(tokenHeaderName string) func(s *state.State, r *http.Request) response.Response {
	return func(s *state.State, r *http.Request) response.Response {
		token := r.Header.Get(tokenHeaderName)
		if token == "" {
			return response.Unauthorized(fmt.Errorf("missing header %q", tokenHeaderName))
		}

		snap := e.provider.Snap()

		nodeToken, err := os.ReadFile(snap.NodeTokenFile())
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to read node access token: %w", err))
		}

		if string(nodeToken) != token {
			return response.Unauthorized(fmt.Errorf("invalid token"))
		}

		return response.EmptySyncResponse
	}
}
