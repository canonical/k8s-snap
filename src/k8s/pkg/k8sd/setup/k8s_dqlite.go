package setup

import (
	"fmt"
	"os"
	"path"

	"github.com/canonical/k8s/pkg/snap"
	"gopkg.in/yaml.v2"
)

type k8sDqliteInit struct {
	Address string   `yaml:"Address,omitempty"`
	Cluster []string `yaml:"Cluster,omitempty"`
}

func K8sDqlite(snap snap.Snap, address string, cluster []string) error {
	b, err := yaml.Marshal(&k8sDqliteInit{Address: address, Cluster: cluster})
	if err != nil {
		return fmt.Errorf("failed to create k8s-dqlite init.yaml file for address=%s cluster=%v: %w", address, cluster, err)
	}

	if err := os.WriteFile(path.Join(snap.K8sDqliteStateDir(), "init.yaml"), b, 0600); err != nil {
		return fmt.Errorf("failed to write k8s-dqlite init.yaml: %w", err)
	}
	return nil
}
