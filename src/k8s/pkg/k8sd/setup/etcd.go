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

type etcdTransportSecurity struct {
	CertFile      string `yaml:"cert-file,omitempty"`
	KeyFile       string `yaml:"key-file,omitempty"`
	TrustedCAFile string `yaml:"trusted-ca-file,omitempty"`
}

type etcdConfig struct {
	Name                     string `yaml:"name,omitempty,omitempty"`
	DataDir                  string `yaml:"data-dir,omitempty"`
	AdvertiseClientURLs      string `yaml:"advertise-client-urls,omitempty"`
	ListenClientURLs         string `yaml:"listen-client-urls,omitempty"`
	ListenPeerURLs           string `yaml:"listen-peer-urls,omitempty"`
	InitialClusterState      string `yaml:"initial-cluster-state,omitempty"`
	InitialCluster           string `yaml:"initial-cluster,omitempty"`
	InitialAdvertisePeerURLs string `yaml:"initial-advertise-peer-urls,omitempty"`

	ClientTransportSecurity etcdTransportSecurity `yaml:"client-transport-security,omitempty"`
	PeerTransportSecurity   etcdTransportSecurity `yaml:"peer-transport-security,omitempty"`
}

type etcdRegisterConfig struct {
	PeerURL       string   `yaml:"peer-url,omitempty"`
	ClientURLs    []string `yaml:"client-urls,omitempty"`
	CertFile      string   `yaml:"cert-file,omitempty"`
	KeyFile       string   `yaml:"key-file,omitempty"`
	TrustedCAFile string   `yaml:"trusted-ca-file,omitempty"`
}

func newEtcdConfig(snap snap.Snap, name, clientURL, peerURL string, clientURLs []string) etcdConfig {
	clusterState := "new"
	if len(clientURLs) > 0 {
		clusterState = "existing"
	}
	return etcdConfig{
		Name:                     name,
		DataDir:                  filepath.Join(snap.K8sDqliteStateDir(), "data"),
		InitialCluster:           fmt.Sprintf("%s=%s", name, peerURL), // NOTE: will be updated for joining nodes
		InitialClusterState:      clusterState,
		InitialAdvertisePeerURLs: peerURL,
		ListenPeerURLs:           peerURL,
		AdvertiseClientURLs:      clientURL,
		ListenClientURLs:         clientURL,
		ClientTransportSecurity: etcdTransportSecurity{
			TrustedCAFile: filepath.Join(snap.EtcdPKIDir(), "ca.crt"),
			CertFile:      filepath.Join(snap.EtcdPKIDir(), "server.crt"),
			KeyFile:       filepath.Join(snap.EtcdPKIDir(), "server.key"),
		},
		PeerTransportSecurity: etcdTransportSecurity{
			TrustedCAFile: filepath.Join(snap.EtcdPKIDir(), "ca.crt"),
			CertFile:      filepath.Join(snap.EtcdPKIDir(), "peer.crt"),
			KeyFile:       filepath.Join(snap.EtcdPKIDir(), "peer.key"),
		},
	}
}

func newEtcdRegisterConfig(snap snap.Snap, peerURL string, clientURLs []string) etcdRegisterConfig {
	return etcdRegisterConfig{
		PeerURL:       peerURL,
		ClientURLs:    clientURLs,
		TrustedCAFile: filepath.Join(snap.EtcdPKIDir(), "ca.crt"),
		CertFile:      filepath.Join(snap.EtcdPKIDir(), "server.crt"),
		KeyFile:       filepath.Join(snap.EtcdPKIDir(), "server.key"),
	}
}

func Etcd(snap snap.Snap, name, clientURL, peerURL string, clientURLs []string, extraArgs map[string]*string) error {
	if b, err := yaml.Marshal(newEtcdConfig(snap, name, clientURL, peerURL, clientURLs)); err != nil {
		return fmt.Errorf("failed to create etcd.yaml file for name=%q address=%q: %w", name, peerURL, err)
	} else if err := os.WriteFile(filepath.Join(snap.K8sDqliteStateDir(), "etcd.yaml"), b, 0600); err != nil {
		return fmt.Errorf("failed to write etcd.yaml config for name=%q address=%q: %w", name, peerURL, err)
	}

	if b, err := yaml.Marshal(newEtcdRegisterConfig(snap, peerURL, clientURLs)); err != nil {
		return fmt.Errorf("failed to create register.yaml file for name=%q address=%q: %w", name, peerURL, err)
	} else if err := os.WriteFile(filepath.Join(snap.K8sDqliteStateDir(), "register.yaml"), b, 0600); err != nil {
		return fmt.Errorf("failed to write register.yaml file for name=%q address=%q: %w", name, peerURL, err)
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "k8s-dqlite", map[string]string{
		"--etcd-mode":   "true",
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
