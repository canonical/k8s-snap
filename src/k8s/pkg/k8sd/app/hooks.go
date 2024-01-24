package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"path"

	"github.com/canonical/k8s/pkg/k8s/setup"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/cert"
	"github.com/canonical/microcluster/state"
)

// onBootstrap is called after we bootstrap the first cluster node.
// onBootstrap configures local services then writes the cluster config on the database.
func onBootstrap(s *state.State, initConfig map[string]string) error {
	snap := snap.SnapFromContext(s.Context)

	err := setup.InitFolders(snap.DataPath("args"))
	if err != nil {
		return fmt.Errorf("failed to setup folders: %w", err)
	}

	err = setup.InitServiceArgs(snap, nil)
	if err != nil {
		return fmt.Errorf("failed to setup service arguments: %w", err)
	}

	err = setup.InitContainerd(snap.Path("k8s/config/containerd/config.toml"), snap.Path("opt/cni/bin/"))
	if err != nil {
		return fmt.Errorf("failed to initialize containerd: %w", err)
	}

	certMan, err := setup.InitCertificates(nil)
	if err != nil {
		return fmt.Errorf("failed to setup certificates: %w", err)
	}

	err = setup.InitKubeconfigs(s.Context, s, certMan.CA, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to kubeconfig files: %w", err)
	}

	err = setup.InitKubeApiserver(snap.Path("k8s/config/apiserver-token-hook.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to initialize kube-apiserver: %w", err)
	}

	err = setup.InitPermissions(s.Context, snap)
	if err != nil {
		return fmt.Errorf("failed to setup permissions: %w", err)
	}

	// TODO(neoaggelos): these should be done with "database.SetClusterConfig()" at the end of the bootstrap
	err = cert.WriteCertKeyPairToK8sd(s.Context, s, "certificates-k8s-dqlite",
		path.Join(cert.K8sDqlitePkiPath, "cluster.crt"), path.Join(cert.K8sDqlitePkiPath, "cluster.key"))
	if err != nil {
		return fmt.Errorf("failed to write k8s-dqlite cert to k8sd: %w", err)
	}
	err = cert.WriteCertKeyPairToK8sd(s.Context, s, "certificates-ca",
		path.Join(cert.KubePkiPath, "ca.crt"), path.Join(cert.KubePkiPath, "ca.key"))
	if err != nil {
		return fmt.Errorf("failed to write CA to k8sd: %w", err)
	}

	// TODO(neoaggelos): configure k8s-dqlite init.yaml file, as it is currently only left to guess for defaults
	//                   - see "k8s::init::k8s_dqlite" in k8s/lib.sh for details.
	//                   - do not bind on 127.0.0.1, use configuration option or fallback to default address like microcluster.

	// TODO(neoaggelos): first generate config then reconcile state
	s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		return database.SetClusterConfig(ctx, tx, database.ClusterConfig{
			Cluster: database.ClusterConfigCluster{
				CIDR: "10.1.0.0/16",
			},
			APIServer: database.ClusterConfigAPIServer{
				RBAC:       "true",
				SecurePort: "6443",
			},
		})
	})

	err = snap.StartService(s.Context, "k8s")
	if err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	return nil
}

// onPostJoin is called when a control plane node joins the cluster.
// onPostJoin retrieves the cluster config from the database and configures local services.
func onPostJoin(s *state.State, initConfig map[string]string) error {
	snap := snap.SnapFromContext(s.Context)

	var clusterConfig database.ClusterConfig
	if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		clusterConfig, err = database.GetClusterConfig(ctx, tx)
		return err
	}); err != nil {
		return fmt.Errorf("failed to retrieve the cluster configuration from the database: %w", err)
	}

	if err := setup.InitFolders(snap.DataPath("args")); err != nil {
		return fmt.Errorf("failed to setup folders: %w", err)
	}

	if err := setup.InitServiceArgs(snap, map[string]map[string]string{
		"kube-apiserver": {
			"--secure-port": clusterConfig.APIServer.SecurePort,
		},
		"kube-proxy": {
			"--cluster-cidr": clusterConfig.Cluster.CIDR,
		},
	}); err != nil {
		return fmt.Errorf("failed to setup service arguments: %w", err)
	}

	if err := setup.InitContainerd(snap.Path("k8s/config/containerd/config.toml"), snap.Path("opt/cni/bin/")); err != nil {
		return fmt.Errorf("failed to initialize containerd: %w", err)
	}

	if err := cert.StoreCertKeyPair(clusterConfig.Certificates.CertificateAuthorityCert, clusterConfig.Certificates.CertificateAuthorityKey, path.Join(cert.KubePkiPath, "ca.crt"), path.Join(cert.KubePkiPath, "ca.key")); err != nil {
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

func onPostRemove(s *state.State, force bool) error {
	// TODO: the current node has left the cluster, stop services and reset configs
	return nil
}
