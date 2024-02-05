package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/cert"
	"github.com/canonical/k8s/pkg/utils/dqlite"
	"github.com/canonical/microcluster/state"
	"gopkg.in/yaml.v2"
)

// JoinK8sDqliteCluster joins a node to an existing k8s-dqlite cluster. It:
//
//   - retrieves k8s-dqlite certificates and address from cluster node (k8sd is already joined at this point so we can access the certificates)
//   - stores new certificates in k8s-dqlite cluster directory
//   - writes k8s-dqlite init file with the cluster node information
//   - starts k8s-dqlite
func JoinK8sDqliteCluster(ctx context.Context, state *state.State, snap snap.Snap) error {
	clusterConfig, err := utils.GetClusterConfig(ctx, state)
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %w", err)
	}

	k8sDqliteCertPair, err := cert.NewCertKeyPairFromPEM([]byte(clusterConfig.Certificates.K8sDqliteCert), []byte(clusterConfig.Certificates.K8sDqliteKey))
	if err != nil {
		return fmt.Errorf("failed to create k8s-dqlite cert from pem: %w", err)
	}

	if err := k8sDqliteCertPair.SaveCertificate(snap.CommonPath(cert.K8sDqlitePkiPath, "cluster.crt")); err != nil {
		return fmt.Errorf("failed to write k8s-dqlite cert: %w", err)
	}
	if err := k8sDqliteCertPair.SavePrivateKey(snap.CommonPath(cert.K8sDqlitePkiPath, "cluster.key")); err != nil {
		return fmt.Errorf("failed to write k8s-dqlite key: %w", err)
	}

	leader, err := state.Leader()
	if err != nil {
		return fmt.Errorf("failed to get dqlite leader: %w", err)
	}

	members, err := leader.GetClusterMembers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dqlite members: %w", err)
	}
	clusterAddrs := make([]string, len(members))

	for _, member := range members {
		clusterAddrs = append(clusterAddrs, fmt.Sprintf("%s:%d", member.Address.Addr(), clusterConfig.K8sDqlite.Port))
	}

	initFile := K8sDqliteInit{
		Cluster: clusterAddrs,
	}
	if err := WriteClusterInitFile(initFile); err != nil {
		return fmt.Errorf("failed to update cluster info.yaml file: %w", err)
	}

	if err := snap.StartService(ctx, "k8s-dqlite"); err != nil {
		return fmt.Errorf("failed to stop k8s-dqlite: %w", err)
	}

	return nil
}

func LeaveK8sDqliteCluster(ctx context.Context, snap snap.Snap, state *state.State) error {
	clusterConfig, err := utils.GetClusterConfig(ctx, state)
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %w", err)
	}

	address := fmt.Sprintf("%s:%d", state.Address().Hostname(), clusterConfig.K8sDqlite.Port)

	members, err := dqlite.GetK8sDqliteClusterMembers(ctx, snap)
	if err != nil {
		return fmt.Errorf("failed to get cluster members: %w", err)
	}

	// TODO: handle case where node is leader but there are successors (e.g. use client.Transfer)
	if err := dqlite.IsLeaderWithoutSuccessor(ctx, members, address); err != nil {
		return fmt.Errorf("failed to leave cluster: %w", err)
	}

	return utils.RunCommand(ctx, snap.Path("k8s/wrappers/commands/dqlite"), "k8s", fmt.Sprintf(".remove %s", address))
}

// K8sDqliteInit represents the yaml file structure of the dqlite `init.yaml` file.
type K8sDqliteInit struct {
	ID      uint64   `yaml:"ID,omitempty"`
	Address string   `yaml:"Address,omitempty"`
	Role    int      `yaml:"Role,omitempty"`
	Cluster []string `yaml:"Cluster,omitempty"`
}

// WriteClusterInitFile writes an `init.yaml` file to the k8s-dqlite directory
// that contains the informations to join an existing cluster (e.g. members addresses)
// and is picked up by k8s-dqlite on startup.
func WriteClusterInitFile(init K8sDqliteInit) error {
	marshaled, err := yaml.Marshal(&init)
	if err != nil {
		return fmt.Errorf("failed to marshal cluster init data: %w", err)
	}

	if err := os.WriteFile(filepath.Join("/var/snap/k8s/common", cert.K8sDqlitePkiPath, "init.yaml"), []byte(marshaled), 0644); err != nil {
		return fmt.Errorf("failed to write init.yaml to %s: %w", cert.K8sDqlitePkiPath, err)
	}
	return nil
}
