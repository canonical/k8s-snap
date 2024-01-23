package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/canonical/microcluster/cluster"
	"github.com/mitchellh/mapstructure"
)

var (
	clusterConfigsStmts = map[string]int{
		"upsert-config": mustPrepareStatement("cluster-configs", "upsert-config.sql"),
		"select-all":    mustPrepareStatement("cluster-configs", "select-all.sql"),
	}
)

type ClusterConfigAPIServer struct {
	// TODO(neoaggelos): change to int, after we resolve this error from mapstructure
	// - 'apiserver-secure-port' expected type 'uint', got unconvertible type 'string', value: '6443'
	SecurePort string `mapstructure:"apiserver-secure-port,omitempty"`
	// TODO(neoaggelos): change to bool, after we resolve this error from mapstructure
	// - 'apiserver-rbac' expected type 'bool', got unconvertible type 'string', value: '1'
	RBAC                string `mapstructure:"apiserver-rbac,omitempty"`
	ServiceAccountKey   string `mapstructure:"apiserver-service-account-key,omitempty"`
	Datastore           string `mapstructure:"apiserver-datastore,omitempty"`
	DatastoreURL        string `mapstructure:"apiserver-datastore-url,omitempty"`
	DatastoreCA         string `mapstructure:"apiserver-datastore-ca,omitempty"`
	DatastoreClientCert string `mapstructure:"apiserver-datastore-client-crt,omitempty"`
	DatastoreClientKey  string `mapstructure:"apiserver-datastore-client-key,omitempty"`
}

type ClusterConfigKubelet struct {
	CloudProvider string `mapstructure:"kubelet-cloud-provider,omitempty"`
	ClusterDNS    string `mapstructure:"kubelet-cluster-dns,omitempty"`
	ClusterDomain string `mapstructure:"kubelet-cluster-domain,omitempty"`
}

type ClusterConfigCertificates struct {
	CertificateAuthorityCert string `mapstructure:"certificates-ca-crt,omitempty"`
	CertificateAuthorityKey  string `mapstructure:"certificates-ca-key,omitempty"`
	APIServerToKubeletCert   string `mapstructure:"certificates-apiserver-to-kubelet-crt,omitempty"`
	APIServerToKubeletKey    string `mapstructure:"certificates-apiserver-to-kubelet-key,omitempty"`
	K8sDqliteCert            string `mapstructure:"certificates-k8s-dqlite-crt,omitempty"`
	K8sDqliteKey             string `mapstructure:"certificates-k8s-dqlite-key,omitempty"`
}

type ClusterConfigCluster struct {
	CIDR string `mapstructure:"cluster-cidr,omitempty"`
}

type ClusterConfig struct {
	Cluster      ClusterConfigCluster      `mapstructure:",squash"`
	Certificates ClusterConfigCertificates `mapstructure:",squash"`
	Kubelet      ClusterConfigKubelet      `mapstructure:",squash"`
	APIServer    ClusterConfigAPIServer    `mapstructure:",squash"`
}

// UpdateClusterConfig inserts or updates a single cluster config entry.
func UpdateClusterConfig(ctx context.Context, tx *sql.Tx, key string, value any) error {
	upsertTxStmt, err := cluster.Stmt(tx, clusterConfigsStmts["upsert-config"])
	if err != nil {
		return fmt.Errorf("failed to prepare upsert statement: %w", err)
	}

	if _, err := upsertTxStmt.ExecContext(ctx, key, value); err != nil {
		return fmt.Errorf("upsert query (write %s->%s) failed: %w", key, value, err)
	}

	return nil
}

// SetClusterConfig inserts or updates a cluster config instance.
// Members that are not set will be emptied.
func SetClusterConfig(ctx context.Context, tx *sql.Tx, clusterConfig ClusterConfig) error {
	configMap := make(map[string]any)
	if err := mapstructure.Decode(clusterConfig, &configMap); err != nil {
		return fmt.Errorf("failed to decode cluster config: %w", err)
	}

	for key, value := range configMap {
		// TODO: Create a multi-update SQL query instead.
		if err := UpdateClusterConfig(ctx, tx, key, value); err != nil {
			return fmt.Errorf("upsert query (write %s->%s) failed: %w", key, value, err)
		}
	}

	return nil
}

// GetClusterConfig retrieves the cluster configuration from the database.
func GetClusterConfig(ctx context.Context, tx *sql.Tx) (ClusterConfig, error) {
	txStmt, err := cluster.Stmt(tx, clusterConfigsStmts["select-all"])
	if err != nil {
		return ClusterConfig{}, fmt.Errorf("failed to prepare statement: %w", err)
	}

	rows, err := txStmt.QueryContext(ctx)
	if err != nil && err != sql.ErrNoRows {
		return ClusterConfig{}, fmt.Errorf("failed to get cluster config: %w", err)
	}

	configMap := make(map[string]any)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return ClusterConfig{}, fmt.Errorf("failed to scan row: %w", err)
		}
		configMap[key] = value
	}

	var clusterConfig ClusterConfig
	if err := mapstructure.Decode(configMap, &clusterConfig); err != nil {
		return ClusterConfig{}, fmt.Errorf("failed to decode cluster config: %w", err)
	}

	return clusterConfig, nil
}
