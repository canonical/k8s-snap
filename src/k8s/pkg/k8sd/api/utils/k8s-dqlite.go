package utils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/rest/types"
	"github.com/canonical/microcluster/state"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	// TODO(bschimke): Do not use global state here.
	clusterDir       = snap.CommonPath("var/lib/k8s-dqlite")
	clusterBackupDir = snap.CommonPath("var/lib/k8s-dqlite-backup")
)

// WriteK8sDqliteCertInfoToK8sd gets local cert and key and stores them in the k8sd database.
func WriteK8sDqliteCertInfoToK8sd(ctx context.Context, state *state.State) error {
	// Read cert and key from local
	cert, err := os.ReadFile(path.Join(clusterDir, "cluster.crt"))
	if err != nil {
		return fmt.Errorf("failed to read cluster.cert from %s: %w", clusterDir, err)
	}
	key, err := os.ReadFile(path.Join(clusterDir, "cluster.key"))
	if err != nil {
		return fmt.Errorf("failed to read cluster.key from %s: %w", clusterDir, err)
	}

	logrus.WithField("cert_length", len(string(cert))).WithField("key_length", len(string(key))).Debug("Writing k8s-dqlite cert and key to database")
	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err = database.CreateCertificate(ctx, tx, "k8s-dqlite", string(cert), string(key))
		if err != nil {
			return fmt.Errorf("failed to write k8s-dqlite certs and key to database: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to perform k8s-dqlite certificate transaction write request: %w", err)
	}
	return nil
}

// UpdateK8sDqlite joins the k8s-dqlite databases by performing the following steps:
//
//	Stop apiserver and k8s-dqlite
//	Backup existing k8s-cluster
//	Retrieve k8s-dqlite certificates from cluster node (k8sd is already joined at this point so we can access the certificates)
//	Store new certificates in k8s-dqlite cluster directory
//	Get current address and port from backup `info.yaml`
//	Get voters from join token (TODO: implement new token that encodes k8sd token and voter info into one token)
//	for now - just use the voter IPs from the k8sd token and assume the default k8s-dqlite port
//	write k8s-dqlite init file
//	restart services and wait to join
func UpdateK8sDqlite(ctx context.Context, state *state.State, voters []string, host string) error {
	logrus.Debug("Stop kube-apiserver")
	if err := snap.StopService(ctx, "kube-apiserver"); err != nil {
		return fmt.Errorf("failed to stop apiserver: %w", err)
	}

	logrus.Debug("Stop k8s-dqlite")
	if err := snap.StopService(ctx, "k8s-dqlite"); err != nil {
		return fmt.Errorf("failed to stop k8s-dqlite: %w", err)
	}

	logrus.Debug("Clean backupdir")
	if err := os.RemoveAll(clusterBackupDir); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	logrus.Debugf("Rename %s to %s", clusterDir, clusterBackupDir)
	if err := os.Rename(clusterDir, clusterBackupDir); err != nil {
		return fmt.Errorf("failed to move directory: %w", err)
	}

	logrus.Debug("Create new cluster dir")
	if err := os.Mkdir(clusterDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	logrus.Debug("Update cluster certificates")
	if err := updateClusterCertificate(ctx, state); err != nil {
		return fmt.Errorf("failed to update k8s-dqlite cluster certificate: %w", err)
	}

	logrus.Debug("Update cluster info file")
	if err := createClusterInitFile(voters, host); err != nil {
		return fmt.Errorf("failed to update cluster info.yaml file: %w", err)
	}

	logrus.Debug("Start k8s-dqlite service")
	if err := snap.StartService(ctx, "k8s-dqlite"); err != nil {
		return fmt.Errorf("failed to stop k8s-dqlite: %w", err)
	}

	logrus.Debug("Wait for node to join")
	if err := waitForNodeJoin(ctx, host); err != nil {
		return fmt.Errorf("failed to wait for k8s-dqlite cluster to join: %w", err)
	}

	logrus.Debug("Start kube-apiserver")
	if err := snap.StartService(ctx, "kube-apiserver"); err != nil {
		return fmt.Errorf("failed to stop apiserver: %w", err)
	}

	return nil
}

// updateClusterCertificate read the k8s-dqlite certificate & key from the k8sd database and write it
// to the joining node k8s-dqlite directory.
func updateClusterCertificate(ctx context.Context, state *state.State) error {
	// Get the certificates from the k8sd cluster
	var cert, key string
	var err error
	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		cert, key, err = database.GetCertificateAndKey(ctx, tx, "k8s-dqlite")
		if err != nil {
			return fmt.Errorf("failed to get k8s-dqlite certs and key from database: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to perform k8s-dqlite certificate transaction request: %w", err)
	}

	logrus.WithField("cert", cert).Debug("Write k8s-dqlite certificate")
	// Write them to the k8s-dqlite cluster directory
	if err := os.WriteFile(path.Join(clusterDir, "cluster.crt"), []byte(cert), 0644); err != nil {
		return fmt.Errorf("failed to write cluster.cert to %s: %w", clusterDir, err)
	}
	logrus.WithField("key", key).Debug("Write k8s-dqlite cert key")
	if err := os.WriteFile(path.Join(clusterDir, "cluster.key"), []byte(key), 0644); err != nil {
		return fmt.Errorf("failed to write cluster.key to %s: %w", clusterDir, err)
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
	// TODO(bschimke): add the port as a configuration option to k8sd so that this can be determined dynamically.
	port := 9000

	// Assumes that all cluster members use the same port for k8s-dqlite
	// TODO: do not reuse voter information from the k8sd token but encode the real k8s-dqlite
	// member data into a new token.
	v := []string{}
	addrPorts, err := types.ParseAddrPorts(voters)
	if err != nil {
		return fmt.Errorf("failed to parse voter addresses: %w", err)
	}
	for _, a := range addrPorts {
		v = append(v, fmt.Sprintf("%s:%d", a.Addr(), port))
	}

	initData := clusterInit{
		Cluster: v,
		Address: fmt.Sprintf("%s:%d", host, port),
	}

	marshaled, err := yaml.Marshal(&initData)
	if err != nil {
		return fmt.Errorf("failed to marshal cluster init data: %w", err)
	}

	if err := os.WriteFile(filepath.Join(clusterDir, "init.yaml"), []byte(marshaled), 0644); err != nil {
		return fmt.Errorf("failed to write init.yaml to %s: %w", clusterDir, err)
	}
	return nil
}

func waitForNodeJoin(ctx context.Context, host string) error {
	ch := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// TODO: Use go-dqlite lib instead of shelling out.
				cmd := exec.Command(
					snap.Path("bin/dqlite"),
					"-s", fmt.Sprintf("file://%s/cluster.yaml", clusterDir),
					"-c", fmt.Sprintf("%s/cluster.crt", clusterDir),
					"-k", fmt.Sprintf("%s/cluster.key", clusterDir),
					"-f", "json", "k8s", ".cluster",
				)

				out, err := cmd.CombinedOutput()
				if err == nil && strings.Contains(string(out), host) {
					ch <- struct{}{}
					return
				}
			}
			time.Sleep(time.Second)
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Minute):
		return fmt.Errorf("Node did not finish joining the cluster within time.")
	case <-ch:
		return nil
	}
}
