package setup

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/cert"
	"github.com/canonical/microcluster/state"
	"gopkg.in/yaml.v2"
)

var (
	// TODO(bschimke): add the port as a configuration option to k8sd so that this can be determined dynamically.
	k8sDqliteDefaultPort = 9000
)

// JoinK8sDqliteCluster joins a node to an existing k8s-dqlite cluster. It:
//
//   - retrieves k8s-dqlite certificates from cluster node (k8sd is already joined at this point so we can access the certificates)
//   - stores new certificates in k8s-dqlite cluster directory
//   - writes k8s-dqlite init file with the cluster node information
func JoinK8sDqliteCluster(ctx context.Context, state *state.State, snap snap.Snap, knownHost string) error {
	// TODO: Cleanup once the cluster config is fully fetched from the database and not from the RPC endpoint above.
	var crt, key string
	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		config, err := database.GetClusterConfig(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to get k8s-dqlite cert and key from database: %w", err)
		}
		crt = config.Certificates.K8sDqliteCert
		key = config.Certificates.K8sDqliteKey
		return nil
	}); err != nil {
		return fmt.Errorf("failed to perform k8s-dqlite transaction request: %w", err)
	}

	if err := cert.StoreCertKeyPair(crt, key, path.Join(cert.K8sDqlitePkiPath, "cluster.crt"), path.Join(cert.K8sDqlitePkiPath, "cluster.key")); err != nil {
		return fmt.Errorf("failed to update k8s-dqlite cluster certificate: %w", err)
	}

	if err := createClusterInitFile(knownHost); err != nil {
		return fmt.Errorf("failed to update cluster info.yaml file: %w", err)
	}

	if err := snap.StartService(ctx, "k8s-dqlite"); err != nil {
		return fmt.Errorf("failed to stop k8s-dqlite: %w", err)
	}

	return nil
}

// clusterInit represents the yaml file structure of the dqlite `init.yaml` file.
type clusterInit struct {
	ID      string   `yaml:"ID,omitempty"`
	Address string   `yaml:"Address,omitempty"`
	Role    int      `yaml:"Role,omitempty"`
	Cluster []string `yaml:"Cluster,omitempty"`
}

// createClusterInitFile writes an `init.yaml` file to the k8s-dqlite directory
// that contains the informations to join an existing cluster (e.g. members addresses)
// and is picked up by k8s-dqlite on startup.
func createClusterInitFile(knownHost string) error {
	// Assumes that all cluster members use the same port for k8s-dqlite
	// TODO: do not reuse voter information from the k8sd token but encode the real k8s-dqlite
	// member data into a new token.
	initData := clusterInit{
		Cluster: []string{fmt.Sprintf("%s:%d", knownHost, k8sDqliteDefaultPort)},
	}

	marshaled, err := yaml.Marshal(&initData)
	if err != nil {
		return fmt.Errorf("failed to marshal cluster init data: %w", err)
	}

	if err := os.WriteFile(filepath.Join(cert.K8sDqlitePkiPath, "init.yaml"), []byte(marshaled), 0644); err != nil {
		return fmt.Errorf("failed to write init.yaml to %s: %w", cert.K8sDqlitePkiPath, err)
	}
	return nil
}
