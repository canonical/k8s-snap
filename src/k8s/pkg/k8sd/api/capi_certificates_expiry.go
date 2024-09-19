package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v3/state"
)

func (e *Endpoints) postCertificatesExpiry(s state.State, r *http.Request) response.Response {
	request := apiv1.ClusterAPICertificatesExpiryRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	config, err := database.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	expiry, err := utils.GetCertExpiry(config.Certificates.GetAdminClientCert())
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get certificate expiry: %w", err))
	}

	return response.SyncResponse(true, &apiv1.ClusterAPICertificatesExpiryResponse{
		Expiry: expiry,
	})
}
