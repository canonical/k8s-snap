package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v3/state"
)

func (e *Endpoints) ValidateNodeTokenAccessHandler(tokenHeaderName string) func(s state.State, r *http.Request) (bool, response.Response) {
	return func(s state.State, r *http.Request) (bool, response.Response) {
		token := r.Header.Get(tokenHeaderName)
		if token == "" {
			return false, response.Unauthorized(fmt.Errorf("missing header %q", tokenHeaderName))
		}

		snap := e.provider.Snap()

		nodeToken, err := os.ReadFile(snap.NodeTokenFile())
		if err != nil {
			return false, response.InternalError(fmt.Errorf("failed to read node access token: %w", err))
		}

		if strings.TrimSpace(string(nodeToken)) != token {
			return false, response.Unauthorized(fmt.Errorf("invalid token"))
		}

		return true, nil
	}
}
