package app

import (
	"fmt"
	"log"
	"path"

	"github.com/canonical/k8s/pkg/k8s/setup"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/cert"
	"github.com/canonical/microcluster/state"
)

// onPostJoin is called when a control plane node joins the cluster.
// onPostJoin retrieves the cluster config from the database and configures local services.
func onPostJoin(s *state.State, initConfig map[string]string) error {
	snap := snap.SnapFromContext(s.Context)

	clusterConfig, err := utils.GetClusterConfig(s.Context, s)
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %w", err)
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

	caKeyPair, err := cert.NewCertKeyPairFromPEM([]byte(clusterConfig.Certificates.CACert), []byte(clusterConfig.Certificates.CAKey))
	if err != nil {
		return fmt.Errorf("failed to create CA from pem: %w", err)
	}

	if err := caKeyPair.SaveCertificate(path.Join(cert.KubePkiPath, "ca.crt")); err != nil {
		return fmt.Errorf("failed to write CA cert: %w", err)
	}
	if err := caKeyPair.SavePrivateKey(path.Join(cert.KubePkiPath, "ca.key")); err != nil {
		return fmt.Errorf("failed to write CA key: %w", err)
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

	if err := setup.JoinK8sDqliteCluster(s.Context, s, snap); err != nil {
		return fmt.Errorf("failed to join k8s-dqlite nodes: %w", err)
	}

	if err := snap.StartService(s.Context, "k8s"); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	return nil
}

func onPreRemove(s *state.State, force bool) error {
	snap := snap.SnapFromContext(s.Context)

	isWorker, err := snap.IsWorker()
	if err != nil {
		return fmt.Errorf("failed to check if node is a worker: %w", err)
	}

	if isWorker {
		return fmt.Errorf("can not run remove-node on workers")
	}

	// Remove k8s dqlite node from cluster.
	// Fails if the k8s-dqlite cluster would not have a leader afterwards.
	log.Println("Leave k8s-dqlite cluster")
	err = setup.LeaveK8sDqliteCluster(s.Context, snap, s)
	if err != nil {
		return fmt.Errorf("failed to leave k8s-dqlite cluster: %w", err)
	}

	// TODO: Remove node from kubernetes

	return nil
}

func onNewMember(s *state.State) error {
	snap := snap.SnapFromContext(s.Context)

	isWorker, err := snap.IsWorker()
	if err != nil {
		return fmt.Errorf("failed to check if node is a worker: %w", err)
	}

	if isWorker {
		return fmt.Errorf("can not run remove-node on workers")
	}

	return nil
}
