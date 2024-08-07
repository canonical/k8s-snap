package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/microcluster"
	"github.com/canonical/microcluster/v2/state"
)

func (e *Endpoints) postClusterJoinTokens(s state.State, r *http.Request) response.Response {
	req := apiv1.GetJoinTokenRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	hostname, err := utils.CleanHostname(req.Name)
	if err != nil {
		return response.BadRequest(fmt.Errorf("invalid hostname %q: %w", req.Name, err))
	}

	var token string
	if req.Worker {
		token, err = getOrCreateWorkerToken(r.Context(), s, hostname)
	} else {
		token, err = getOrCreateJoinToken(r.Context(), e.provider.MicroCluster(), hostname)
	}
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create token: %w", err))
	}

	return response.SyncResponse(true, &apiv1.GetJoinTokenResponse{EncodedToken: token})
}

func getOrCreateJoinToken(ctx context.Context, m *microcluster.MicroCluster, tokenName string) (string, error) {
	// grab token if it exists and return it
	records, err := m.ListJoinTokens(ctx)
	if err != nil {
		fmt.Println("Failed to get existing tokens. Trying to create a new token.")
	} else {
		for _, record := range records {
			if record.Name == tokenName {
				return record.Token, nil
			}
		}
		fmt.Println("No token exists yet. Creating a new token.")
	}

	// if token does not exist, create a new one
	// TODO(ben): make token expiry configurable
	token, err := m.NewJoinToken(ctx, tokenName, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate a new microcluster join token: %w", err)
	}
	return token, nil
}

func getOrCreateWorkerToken(ctx context.Context, s state.State, nodeName string) (string, error) {
	var token string
	if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		token, err = database.GetOrCreateWorkerNodeToken(ctx, tx, nodeName)
		if err != nil {
			return fmt.Errorf("failed to create worker node token: %w", err)
		}
		return err
	}); err != nil {
		return "", fmt.Errorf("database transaction failed: %w", err)
	}

	remoteAddresses := s.Remotes().Addresses()
	addresses := make([]string, 0, len(remoteAddresses))
	for _, addrPort := range remoteAddresses {
		addresses = append(addresses, addrPort.String())
	}

	cert, err := s.ClusterCert().PublicKeyX509()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster certificate: %w", err)
	}

	info := &types.InternalWorkerNodeToken{
		Secret:        token,
		JoinAddresses: addresses,
		Fingerprint:   utils.CertFingerprint(cert),
	}

	token, err = info.Encode()
	if err != nil {
		return "", fmt.Errorf("failed to encode join token: %w", err)
	}

	return token, nil
}
