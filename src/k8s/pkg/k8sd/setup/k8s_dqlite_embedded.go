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

type k8sDqliteEmbeddedYaml struct {
	Name                     string `yaml:"name,omitempty,omitempty"`
	DataDir                  string `yaml:"data-dir,omitempty"`
	AdvertiseClientURLs      string `yaml:"advertise-client-urls,omitempty"`
	ListenClientURLs         string `yaml:"listen-client-urls,omitempty"`
	ListenPeerURLs           string `yaml:"listen-peer-urls,omitempty"`
	InitialClusterState      string `yaml:"initial-cluster-state,omitempty"`
	InitialCluster           string `yaml:"initial-cluster,omitempty"`
	InitialAdvertisePeerURLs string `yaml:"initial-advertise-peer-urls,omitempty"`
}

type k8sdDqliteEmbeddedConfigYaml struct {
	ClientURLs   []string `yaml:"client-urls,omitempty"`
	PeerURL      string   `yaml:"peer-url,omitempty"`
	CAFile       string   `yaml:"ca-file,omitempty"`
	CertFile     string   `yaml:"cert-file,omitempty"`
	KeyFile      string   `yaml:"key-file,omitempty"`
	PeerCAFile   string   `yaml:"peer-ca-file,omitempty"`
	PeerCertFile string   `yaml:"peer-cert-file,omitempty"`
	PeerKeyFile  string   `yaml:"peer-key-file,omitempty"`
}

func K8sDqliteEmbedded(snap snap.Snap, name, clientURL, peerURL string, clientURLs []string, extraArgs map[string]*string) error {
	clusterState := "new"
	if len(clientURLs) > 0 {
		clusterState = "existing"
	}

	if b, err := yaml.Marshal(&k8sDqliteEmbeddedYaml{
		Name:                     name,
		DataDir:                  filepath.Join(snap.K8sDqliteStateDir(), "data"),
		InitialCluster:           fmt.Sprintf("%s=%s", name, peerURL), // NOTE: will be updated for joining nodes
		InitialClusterState:      clusterState,
		InitialAdvertisePeerURLs: peerURL,
		ListenPeerURLs:           peerURL,
		AdvertiseClientURLs:      clientURL,
		ListenClientURLs:         clientURL,
	}); err != nil {
		return fmt.Errorf("failed to create embedded.yaml file for name=%q address=%q: %w", name, peerURL, err)
	} else if err := os.WriteFile(filepath.Join(snap.K8sDqliteStateDir(), "embedded.yaml"), b, 0600); err != nil {
		return fmt.Errorf("failed to write embedded.yaml config for name=%q address=%q: %w", name, peerURL, err)
	}

	if b, err := yaml.Marshal(&k8sdDqliteEmbeddedConfigYaml{
		ClientURLs:   clientURLs,
		PeerURL:      peerURL,
		CAFile:       filepath.Join(snap.EtcdPKIDir(), "ca.crt"),
		CertFile:     filepath.Join(snap.EtcdPKIDir(), "server.crt"),
		KeyFile:      filepath.Join(snap.EtcdPKIDir(), "server.key"),
		PeerCAFile:   filepath.Join(snap.EtcdPKIDir(), "ca.crt"),
		PeerCertFile: filepath.Join(snap.EtcdPKIDir(), "peer.crt"),
		PeerKeyFile:  filepath.Join(snap.EtcdPKIDir(), "peer.key"),
	}); err != nil {
		return fmt.Errorf("failed to create config.yaml file for name=%q address=%q: %w", name, peerURL, err)
	} else if err := os.WriteFile(filepath.Join(snap.K8sDqliteStateDir(), "config.yaml"), b, 0600); err != nil {
		return fmt.Errorf("failed to write config.yaml file for name=%q address=%q: %w", name, peerURL, err)
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "k8s-dqlite", map[string]string{
		"--embedded":    "true",
		"--storage-dir": snap.K8sDqliteStateDir(),
	}, nil); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}

	// Apply extra arguments after the defaults, so they can override them.
	updateArgs, deleteArgs := utils.ServiceArgsFromMap(extraArgs)
	if _, err := snaputil.UpdateServiceArguments(snap, "k8s-dqlite", updateArgs, deleteArgs); err != nil {
		return fmt.Errorf("failed to write extra arguments: %w", err)
	}
	return nil
}
