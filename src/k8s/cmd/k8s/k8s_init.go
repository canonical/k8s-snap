package k8s

import (
	"fmt"
	"os/exec"

	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/k8s/pkg/k8s/setup"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize the k8s node",
		Long:  "Initialize the necessary folders, permissions, service arguments, certificates and start up the Kubernetes services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.logDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			err := setup.InitFolders()
			if err != nil {
				return fmt.Errorf("failed to setup folders: %w", err)
			}

			err = setup.InitServiceArgs()
			if err != nil {
				return fmt.Errorf("failed to setup service arguments: %w", err)
			}

			err = setup.InitContainerd()
			if err != nil {
				return fmt.Errorf("failed to initialize containerd: %w", err)
			}

			client, err := setup.InitK8sd(cmd.Context(), client.ClusterOpts{
				RemoteAddress: clusterCmdOpts.remoteAddress,
				Debug:         rootCmdOpts.logDebug,
				Port:          clusterCmdOpts.port,
				StorageDir:    clusterCmdOpts.storageDir,
				Verbose:       rootCmdOpts.logVerbose,
			})
			if err != nil {
				return fmt.Errorf("failed to initialize k8sd: %w", err)
			}

			certMan, err := setup.InitCertificates()
			if err != nil {
				return fmt.Errorf("failed to setup certificates: %w", err)
			}

			err = setup.InitKubeconfigs(cmd.Context(), client, certMan.CA)
			if err != nil {
				return fmt.Errorf("failed to kubeconfig files: %w", err)
			}

			err = setup.InitKubeApiserver()
			if err != nil {
				return fmt.Errorf("failed to initialize kube-apiserver: %w", err)
			}

			err = setup.InitPermissions()
			if err != nil {
				return fmt.Errorf("failed to setup permissions: %w", err)
			}

			startCmd := exec.Command("snapctl", "start", "k8s")

			_, err = startCmd.Output()
			if err != nil {
				return fmt.Errorf("failed to start services: %w", err)
			}

			logrus.Infof("Successfully initialized k8s node.")
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
}
