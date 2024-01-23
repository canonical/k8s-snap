package impl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/cert"
	"github.com/canonical/microcluster/rest/types"
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
func JoinK8sDqliteCluster(ctx context.Context, state *state.State, snap snap.Snap, voters []string, host string) error {
	if err := cert.StoreCertKeyPair(ctx, state, "k8s-dqlite", path.Join(cert.K8sDqlitePkiPath, "cluster.crt"), path.Join(cert.K8sDqlitePkiPath, "cluster.key")); err != nil {
		return fmt.Errorf("failed to update k8s-dqlite cluster certificate: %w", err)
	}

	if err := createClusterInitFile(voters, host); err != nil {
		return fmt.Errorf("failed to update cluster info.yaml file: %w", err)
	}

	if err := snap.StartService(ctx, "k8s-dqlite"); err != nil {
		return fmt.Errorf("failed to stop k8s-dqlite: %w", err)
	}

	if err := waitForNodeJoin(ctx, snap.Path("bin/dqlite"), host); err != nil {
		return fmt.Errorf("failed to wait for k8s-dqlite cluster to join: %w", err)
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
func createClusterInitFile(voters []string, host string) error {

	// Assumes that all cluster members use the same port for k8s-dqlite
	// TODO: do not reuse voter information from the k8sd token but encode the real k8s-dqlite
	// member data into a new token.
	v := []string{}
	addrPorts, err := types.ParseAddrPorts(voters)
	if err != nil {
		return fmt.Errorf("failed to parse voter addresses: %w", err)
	}
	for _, a := range addrPorts {
		v = append(v, fmt.Sprintf("%s:%d", a.Addr(), k8sDqliteDefaultPort))
	}

	initData := clusterInit{
		Cluster: v,
		Address: fmt.Sprintf("%s:%d", host, k8sDqliteDefaultPort),
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

func waitForNodeJoin(ctx context.Context, dqlitePath string, host string) error {
	err := utils.WaitUntilReady(ctx, func() (bool, error) {
		cmd := exec.Command(
			dqlitePath,
			"-s", fmt.Sprintf("file://%s/cluster.yaml", cert.K8sDqlitePkiPath),
			"-c", fmt.Sprintf("%s/cluster.crt", cert.K8sDqlitePkiPath),
			"-k", fmt.Sprintf("%s/cluster.key", cert.K8sDqlitePkiPath),
			"-f", "json", "k8s", ".cluster",
		)
		fmt.Println(cmd.Args)
		out, err := cmd.CombinedOutput()
		// dqlite will throw an error if the cluster is not ready yet.
		// We want to retry this case, so the error is not returned in the error result.
		return strings.Contains(string(out), host) && err == nil, nil
	})
	if err != nil {
		return fmt.Errorf("node (%s) did not finish joining the cluster within time: %w", host, err)
	}
	return nil
}
