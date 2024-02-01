package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"path"

	"github.com/canonical/k8s/pkg/k8s/setup"
	"github.com/canonical/k8s/pkg/k8sd/database/clusterconfigs"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/cert"
	"github.com/canonical/microcluster/state"
)

// onPostJoin is called when a control plane node joins the cluster.
// onPostJoin retrieves the cluster config from the database and configures local services.
func onPostJoin(s *state.State, initConfig map[string]string) error {
	snap := snap.SnapFromContext(s.Context)

	var clusterConfig clusterconfigs.ClusterConfig
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		clusterConfig, err = clusterconfigs.GetClusterConfig(ctx, tx)
		return err
	}); err != nil {
		return fmt.Errorf("failed to retrieve the cluster configuration from the database: %w", err)
	}

	if err := setup.InitFolders(snap.DataPath("args")); err != nil {
		return fmt.Errorf("failed to setup folders: %w", err)
	}

	if err := setup.InitServiceArgs(snap, map[string]map[string]string{
		"kube-apiserver": {
			"--secure-port": fmt.Sprintf("%d", clusterConfig.APIServer.SecurePort),
		},
		"kube-proxy": {
			"--cluster-cidr": clusterConfig.Cluster.CIDR,
		},
	}); err != nil {
		return fmt.Errorf("failed to setup service arguments: %w", err)
	}

	if err := setup.InitContainerd(snap); err != nil {
		return fmt.Errorf("failed to initialize containerd: %w", err)
	}

	if err := cert.StoreCertKeyPair(clusterConfig.Certificates.CACert, clusterConfig.Certificates.CAKey, path.Join(cert.KubePkiPath, "ca.crt"), path.Join(cert.KubePkiPath, "ca.key")); err != nil {
		return fmt.Errorf("failed to store CA certificate: %w", err)
	}

	// Use the CA from the cluster to sign the certificates
	caKeyPair, err := cert.LoadCertKeyPair(path.Join(cert.KubePkiPath, "ca.key"), path.Join(cert.KubePkiPath, "ca.crt"))
	if err != nil {
		return fmt.Errorf("failed to read CA: %w", err)
	}
	certMan, err := setup.InitCertificates(caKeyPair)
	if err != nil {
		return fmt.Errorf("failed to setup certificates: %w", err)
	}

	if err := setup.InitKubeconfigs(s.Context, s, certMan.CA, nil, &clusterConfig.APIServer.SecurePort); err != nil {
		return fmt.Errorf("failed to generate kubeconfig files: %w", err)
	}

	if err := setup.InitKubeApiserver(snap.Path("k8s/config/apiserver-token-hook.tmpl")); err != nil {
		return fmt.Errorf("failed to initialize kube-apiserver: %w", err)
	}

	if err := setup.InitPermissions(s.Context, snap); err != nil {
		return fmt.Errorf("failed to setup permissions: %w", err)
	}
	leader, err := s.Leader()
	if err != nil {
		return fmt.Errorf("failed to get dqlite leader: %w", err)
	}

	// TODO(neoaggelos): k8s-dqlite cluster host and port must come from the cluster config.
	host, _, _ := net.SplitHostPort(leader.URL().URL.Host)
	if err := setup.JoinK8sDqliteCluster(s.Context, s, snap, host); err != nil {
		return fmt.Errorf("failed to join k8s-dqlite nodes: %w", err)
	}

	if err := snap.StartService(s.Context, "k8s"); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	return nil
}

func onPreRemove(s *state.State, force bool) error {
	snap := snap.SnapFromContext(s.Context)

	// Remove k8s dqlite node from cluster.
	// Fails if the k8s-dqlite cluster would not have a leader afterwards.
	log.Println("Leave k8s-dqlite cluster")
	err := setup.LeaveK8sDqliteCluster(s.Context, snap, s.Address().Hostname())
	if err != nil {
		return fmt.Errorf("failed to leave k8s-dqlite cluster: %w", err)
	}

	return nil
}
