package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func postWorkerToken(s *state.State, r *http.Request) response.Response {
	var token string
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		token, err = database.GetOrCreateWorkerNodeToken(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to create worker node token: %w", err)
		}
		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("database transaction failed: %w", err))
	}

	remoteAddresses := s.Remotes().Addresses()
	addresses := make([]string, 0, len(remoteAddresses))
	for _, addrPort := range remoteAddresses {
		addresses = append(addresses, addrPort.String())
	}

	info := &types.InternalWorkerNodeToken{
		Token:         token,
		JoinAddresses: addresses,
	}
	token, err := info.Encode()
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to encode join token: %w", err))
	}

	return response.SyncResponse(true, &apiv1.WorkerNodeTokenResponse{EncodedToken: token})
}

func postWorkerInfo(s *state.State, r *http.Request) response.Response {
	// TODO: move authentication through the HTTP token to an AccessHandler for the endpoint.
	token := r.Header.Get("k8sd-token")
	if token == "" {
		return response.Unauthorized(fmt.Errorf("invalid token"))
	}
	var tokenIsValid bool
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		tokenIsValid, err = database.CheckWorkerNodeToken(ctx, tx, token)
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

	req := apiv1.WorkerNodeInfoRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}
	nodeName := req.Hostname
	if nodeName == "" {
		return response.BadRequest(fmt.Errorf("node name cannot be empty"))
	}

	var clusterConfig database.ClusterConfig
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		clusterConfig, err = database.GetClusterConfig(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to retrieve cluster configuration: %w", err)
		}
		return nil
	}); err != nil {
		return response.InternalError(fmt.Errorf("get cluster config database transaction failed: %w", err))
	}

	client, err := k8s.NewClient()
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create kubernetes client: %w", err))
	}
	servers, err := k8s.GetKubeAPIServerEndpoints(s.Context, client)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to retrieve list of known kube-apiserver endpoints: %w", err))
	}

	var (
		kubeletToken string
		proxyToken   string
	)
	for _, i := range []struct {
		token    *string
		name     string
		username string
		groups   []string
	}{
		{token: &kubeletToken, name: "kubelet", username: fmt.Sprintf("system:node:%s", nodeName), groups: []string{"system:nodes"}},
		{token: &proxyToken, name: "kube-proxy", username: "system:kube-proxy"},
	} {
		if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
			t, err := database.GetOrCreateToken(ctx, tx, i.username, i.groups)
			if err != nil {
				return fmt.Errorf("failed to generate %s token for node %q: %w", i.name, nodeName, err)
			}
			*i.token = t
			return nil
		}); err != nil {
			return response.InternalError(fmt.Errorf("create token transaction failed: %w", err))
		}
	}

	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		return database.AddWorkerNode(ctx, tx, nodeName)
	}); err != nil {
		return response.InternalError(fmt.Errorf("add worker node transaction failed: %w", err))
	}

	return response.SyncResponse(true, &apiv1.WorkerNodeInfoResponse{
		CA:             clusterConfig.Certificates.CACert,
		APIServers:     servers,
		ClusterCIDR:    clusterConfig.Cluster.CIDR,
		KubeletToken:   kubeletToken,
		KubeProxyToken: proxyToken,
		ClusterDomain:  clusterConfig.Kubelet.ClusterDomain,
		ClusterDNS:     clusterConfig.Kubelet.ClusterDNS,
		CloudProvider:  clusterConfig.Kubelet.CloudProvider,
	})
}
