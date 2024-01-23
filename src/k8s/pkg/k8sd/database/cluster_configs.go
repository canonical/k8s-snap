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

type ClusterConfig struct {
	// K8sCertificateAuthority represents the Kubernetes Certificate Authority certificate.
	// Empty if we don't use self-signed certificates.
	K8sCertificateAuthority string `mapstructure:"k8s-ca-crt,omitempty"`
	// K8sCertificateAuthorityKey represents the Kubernetes Certificate Authority private key.
	// Empty if we don't use self-signed certificates.
	K8sCertificateAuthorityKey string `mapstructure:"k8s-ca-key,omitempty"`
	// K8sDqliteCertificate represents the Kubernetes Dqlite Certificate.
	K8sDqliteCertificate string `mapstructure:"k8s-dqlite-crt,omitempty"`
	// K8sDqliteKey represents the Kubernetes Dqlite private key.
	K8sDqliteKey string `mapstructure:"k8s-dqlite-key,omitempty"`
	// K8sClusterCIDR represents the Kubernetes Cluster CIDR.
	K8sClusterCIDR string `mapstructure:"k8s-cluster-cidr,omitempty"`
	// KubeletClusterDNS represents the DNS address for the Kubelet in the cluster.
	KubeletClusterDNS string `mapstructure:"kubelet-cluster-dns,omitempty"`
	// KubeletClusterDomain represents the domain for the Kubelet in the cluster.
	KubeletClusterDomain string `mapstructure:"kubelet-cluster-domain,omitempty"`
	// KubeletCloudProvider represents the cloud provider for the Kubelet.
	KubeletCloudProvider string `mapstructure:"kubelet-cloud-provider,omitempty"`
	// APIServerRBAC defines if RBAC (Role-Based Access Control) is enabled for this cluster.
	APIServerRBAC bool `mapstructure:"apiserver-rbac,omitempty"`
	// APIServerDatastore represents the data store configuration for the API server.
	// "k8s-dqlite" by default, "etcd" if using a custom datastore.
	APIServerDatastore string `mapstructure:"apiserver-datastore,omitempty"`
	// APIServerEtcdURL represents the URL for the external etcd service used by the API server.
	APIServerEtcdURL string `mapstructure:"apiserver-etcd-url,omitempty"`
	// APIServerEtcdCertificateAuthority represents the Certificate Authority for external etcd used by the API server.
	APIServerEtcdCertificateAuthority string `mapstructure:"apiserver-etcd-ca,omitempty"`
	// APIServerEtcdClientCertificate represents the client certificate for external etcd used by the API server.
	APIServerEtcdClientCertificate string `mapstructure:"apiserver-etcd-client-crt,omitempty"`
	// APIServerEtcdClientKey represents the client private key for external etcd used by the API server.
	APIServerEtcdClientKey string `mapstructure:"apiserver-etcd-client-key,omitempty"`
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
