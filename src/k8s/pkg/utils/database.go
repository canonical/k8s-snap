package utils

import (
	"context"
	"database/sql"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/canonical/microcluster/state"
)

// GetClusterConfig is a convenience wrapper around the database call to get the cluster config.
func GetClusterConfig(ctx context.Context, state *state.State) (types.ClusterConfig, error) {
	var clusterConfig types.ClusterConfig
	var err error

	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		clusterConfig, err = database.GetClusterConfig(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to get cluster config from database: %w", err)
		}
		return nil
	}); err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to perform cluster config transaction request: %w", err)
	}

	return clusterConfig, nil
}

// GetUserFacingClusterConfig returns the public cluster config.
func GetUserFacingClusterConfig(ctx context.Context, state *state.State) (apiv1.UserFacingClusterConfig, error) {
	cfg, err := GetClusterConfig(ctx, state)
	if err != nil {
		return apiv1.UserFacingClusterConfig{}, fmt.Errorf("failed to get cluster config: %w", err)
	}

	userFacing := apiv1.UserFacingClusterConfig{
		Network: &apiv1.NetworkConfig{
			Enabled: vals.Pointer(true),
		},
		DNS: &apiv1.DNSConfig{
			Enabled:             vals.Pointer(true),
			UpstreamNameservers: cfg.DNS.UpstreamNameservers,
			ServiceIP:           cfg.Kubelet.ClusterDNS,
			ClusterDomain:       cfg.Kubelet.ClusterDomain,
		},
		Ingress: &apiv1.IngressConfig{
			Enabled:             vals.Pointer(false),
			DefaultTLSSecret:    cfg.Ingress.DefaultTLSSecret,
			EnableProxyProtocol: vals.Pointer(false),
		},
		LoadBalancer: &apiv1.LoadBalancerConfig{
			Enabled:        vals.Pointer(false),
			CIDRs:          cfg.LoadBalancer.CIDRs,
			L2Enabled:      vals.Pointer(false),
			L2Interfaces:   cfg.LoadBalancer.L2Interfaces,
			BGPEnabled:     vals.Pointer(false),
			BGPLocalASN:    cfg.LoadBalancer.BGPLocalASN,
			BGPPeerAddress: cfg.LoadBalancer.BGPPeerAddress,
			BGPPeerASN:     cfg.LoadBalancer.BGPPeerASN,
			BGPPeerPort:    cfg.LoadBalancer.BGPPeerPort,
		},
		LocalStorage: &apiv1.LocalStorageConfig{
			Enabled:       vals.Pointer(false),
			LocalPath:     cfg.LocalStorage.LocalPath,
			ReclaimPolicy: cfg.LocalStorage.ReclaimPolicy,
			SetDefault:    vals.Pointer(true),
		},
		Gateway: &apiv1.GatewayConfig{
			Enabled: vals.Pointer(false),
		},
		MetricsServer: &apiv1.MetricsServerConfig{
			Enabled: vals.Pointer(false),
		},
	}

	if cfg.Network.Enabled != nil {
		userFacing.Network.Enabled = cfg.Network.Enabled
	}

	if cfg.DNS.Enabled != nil {
		userFacing.DNS.Enabled = cfg.DNS.Enabled
	}

	if cfg.Ingress.Enabled != nil {
		userFacing.Ingress.Enabled = cfg.Ingress.Enabled
	}

	if cfg.LoadBalancer.Enabled != nil {
		userFacing.LoadBalancer.Enabled = cfg.LoadBalancer.Enabled
	}

	if cfg.LocalStorage.Enabled != nil {
		userFacing.LocalStorage.Enabled = cfg.LocalStorage.Enabled
	}

	if cfg.Gateway.Enabled != nil {
		userFacing.Gateway.Enabled = cfg.Gateway.Enabled
	}

	if cfg.MetricsServer.Enabled != nil {
		userFacing.MetricsServer.Enabled = cfg.MetricsServer.Enabled
	}

	if cfg.Ingress.EnableProxyProtocol != nil {
		userFacing.Ingress.EnableProxyProtocol = cfg.Ingress.EnableProxyProtocol
	}

	if cfg.LoadBalancer.L2Enabled != nil {
		userFacing.LoadBalancer.L2Enabled = cfg.LoadBalancer.L2Enabled
	}

	if cfg.LoadBalancer.BGPEnabled != nil {
		userFacing.LoadBalancer.BGPEnabled = cfg.LoadBalancer.BGPEnabled
	}

	if cfg.LocalStorage.SetDefault != nil {
		userFacing.LocalStorage.SetDefault = cfg.LocalStorage.SetDefault
	}
	return userFacing, nil
}

// CheckWorkerExists is a convenience wrapper around the database call to check if a worker node entry exists.
func CheckWorkerExists(ctx context.Context, state *state.State, name string) (bool, error) {
	var exists bool
	var err error

	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		exists, err = database.CheckWorkerExists(ctx, tx, name)
		if err != nil {
			return fmt.Errorf("failed to get worker node from database: %w", err)
		}
		return nil
	}); err != nil {
		return false, fmt.Errorf("failed to perform check worker node transaction request: %w", err)
	}

	return exists, nil
}

// GetWorkerNodes is a convenience wrapper around the database call to get the worker node names.
func GetWorkerNodes(ctx context.Context, state *state.State) ([]string, error) {
	var workerNodes []string
	var err error

	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		workerNodes, err = database.ListWorkerNodes(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to list worker nodes from database: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to perform list worker nodes transaction request: %w", err)
	}

	return workerNodes, nil
}

// DeleteWorkerNodeEntry is a convenience wrapper around the database call to delete the worker node entry.
func DeleteWorkerNodeEntry(ctx context.Context, state *state.State, name string) error {
	var err error

	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err = database.DeleteWorkerNode(ctx, tx, name)
		if err != nil {
			return fmt.Errorf("failed to delete worker node from database: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to perform delete worker node transaction request: %w", err)
	}

	return nil
}
