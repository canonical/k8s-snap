package setup

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"gopkg.in/yaml.v2"
)

type k8sDqliteEmbeddedInit struct {
	Name               string   `yaml:"Name,omitempty"`
	DataDir            string   `yaml:"DataDir,omitempty"`
	ClientURL          string   `yaml:"ClientURL,omitempty"`
	PeerURL            string   `yaml:"PeerURL,omitempty"`
	ExistingClientURLs []string `yaml:"ExistingClientURLs,omitempty"`
	ExistingPeerURLs   []string `yaml:"ExistingPeerURLs,omitempty"`
	CACertFile         string   `yaml:"CACertFile,omitempty"`
	ServerCertFile     string   `yaml:"ServerCertFile,omitempty"`
	ServerKeyFile      string   `yaml:"ServerKeyFile,omitempty"`
	PeerCertFile       string   `yaml:"PeerCertFile,omitempty"`
	PeerKeyFile        string   `yaml:"PeerKeyFile,omitempty"`
}

func K8sDqliteEmbedded(snap snap.Snap, name string, clientURL, peerURL string, clientURLs, peerURLs []string) error {
	b, err := yaml.Marshal(&k8sDqliteEmbeddedInit{
		Name:               name,
		DataDir:            filepath.Join(snap.K8sDqliteStateDir(), "embedded"),
		ClientURL:          clientURL,
		PeerURL:            peerURL,
		ExistingClientURLs: clientURLs,
		ExistingPeerURLs:   peerURLs,
		CACertFile:         filepath.Join(snap.EtcdPKIDir(), "ca.crt"),
		ServerCertFile:     filepath.Join(snap.EtcdPKIDir(), "server.crt"),
		ServerKeyFile:      filepath.Join(snap.EtcdPKIDir(), "server.key"),
		PeerCertFile:       filepath.Join(snap.EtcdPKIDir(), "peer.crt"),
		PeerKeyFile:        filepath.Join(snap.EtcdPKIDir(), "peer.key"),
	})
	if err != nil {
		return fmt.Errorf("failed to create init.yaml file for name=%s address=%s cluster=%v: %w", name, clientURL, clientURLs, err)
	}

	if err := os.WriteFile(path.Join(snap.K8sDqliteStateDir(), "init.yaml"), b, 0600); err != nil {
		return fmt.Errorf("failed to write init.yaml: %w", err)
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "k8s-dqlite", map[string]string{
		"--mode":        "embedded",
		"--storage-dir": snap.K8sDqliteStateDir(),
	}, nil); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}
	return nil
}
