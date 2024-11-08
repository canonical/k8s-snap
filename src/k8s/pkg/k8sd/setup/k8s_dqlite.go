package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"gopkg.in/yaml.v2"
)

type k8sDqliteInit struct {
	Address string   `yaml:"Address,omitempty"`
	Cluster []string `yaml:"Cluster,omitempty"`
}

func K8sDqlite(snap snap.Snap, address string, cluster []string, extraArgs map[string]*string) error {
	// cleanup in case of existing cluster
	if _, err := os.Stat(filepath.Join(snap.K8sDqliteStateDir(), "cluster.yaml")); err == nil {
		if err := os.RemoveAll(snap.K8sDqliteStateDir()); err != nil {
			return fmt.Errorf("failed to cleanup not-empty k8s-dqlite directory: %w", err)
		}
		if err := os.MkdirAll(snap.K8sDqliteStateDir(), 0o700); err != nil {
			return fmt.Errorf("failed to create k8s-dqlite state directory: %w", err)
		}
	}

	b, err := yaml.Marshal(&k8sDqliteInit{Address: address, Cluster: cluster})
	if err != nil {
		return fmt.Errorf("failed to create init.yaml file for address=%s cluster=%v: %w", address, cluster, err)
	}

	if err := utils.WriteFile(filepath.Join(snap.K8sDqliteStateDir(), "init.yaml"), b, 0o600); err != nil {
		return fmt.Errorf("failed to write init.yaml: %w", err)
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "k8s-dqlite", map[string]string{
		"--listen":      fmt.Sprintf("unix://%s", filepath.Join(snap.K8sDqliteStateDir(), "k8s-dqlite.sock")),
		"--storage-dir": snap.K8sDqliteStateDir(),
	}, nil); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}

	// Apply extra arguments after the defaults, so they can override them.
	updateArgs, deleteArgs := utils.ServiceArgsFromMap(extraArgs)
	if _, err := snaputil.UpdateServiceArguments(snap, "k8s-dqlite", updateArgs, deleteArgs); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}
	return nil
}
