package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) postClusterJoin(s *state.State, r *http.Request) response.Response {
	req := apiv1.JoinClusterRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	hostname, err := utils.CleanHostname(req.Name)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", req.Name, err))
	}

	internalToken := types.InternalWorkerNodeToken{}
	// Check if token is worker token
	if internalToken.Decode(req.Token) == nil {
		// Check Server Auth
		// Get remote certificate from the cluster member
		cert, err := utils.GetRemoteCertificate(req.Address)
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to get certificate of cluster member: %w", err))
		}

		// verify that the fingerprint of the certificate matches the fingerprint of the token
		fingerprint := utils.CertFingerprint(cert)
		if fingerprint != internalToken.Fingerprint {
			return response.BadRequest(fmt.Errorf("server authentication failed: join token fingerprint does not match that of the cluster member"))
		}

		// TODO: add certificate to trusted certificates

		// valid worker node token - let's join the cluster
		// The validation of the token is done when fetching the cluster information.
		if err := e.provider.MicroCluster().NewCluster(hostname, req.Address, map[string]string{"workerToken": req.Token}, 0); err != nil {
			return response.InternalError(fmt.Errorf("failed to join k8sd cluster as worker: %w", err))
		}
	} else {
		// Is not a worker token. let microcluster check if it is a valid control-plane token.
		if err := e.provider.MicroCluster().JoinCluster(hostname, req.Address, req.Token, nil, 0); err != nil {
			return response.InternalError(fmt.Errorf("failed to join k8sd cluster as control plane: %w", err))
		}
	}

	return response.SyncResponse(true, nil)
}
