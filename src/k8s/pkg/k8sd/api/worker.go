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
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

var k8sdWorkerToken = rest.Endpoint{
	Path: "k8sd/worker/token",
	Post: rest.EndpointAction{Handler: k8sdWorkerTokenPost},
}

func k8sdWorkerTokenPost(s *state.State, r *http.Request) response.Response {
	req := apiv1.WorkerNodeJoinRequest{}
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
		return response.InternalError(fmt.Errorf("database transaction failed: %w", err))
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
			return response.InternalError(fmt.Errorf("transaction failed: %w", err))
		}
	}

	token := &types.WorkerNodeToken{
		CA:             clusterConfig.Certificates.CACert,
		APIServers:     servers,
		ClusterCIDR:    clusterConfig.Cluster.CIDR,
		KubeletToken:   kubeletToken,
		KubeProxyToken: proxyToken,
		ClusterDomain:  clusterConfig.Kubelet.ClusterDomain,
		ClusterDNS:     clusterConfig.Kubelet.ClusterDNS,
		CloudProvider:  clusterConfig.Kubelet.CloudProvider,
	}
	encoded, err := token.Encode()
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to encode worker token: %w", err))
	}

	return response.SyncResponse(true, &apiv1.WorkerNodeJoinResponse{
		EncodedToken: encoded,
	})
}
