package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/client"
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

			c, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				StorageDir:    clusterCmdOpts.storageDir,
				RemoteAddress: clusterCmdOpts.remoteAddress,
				Port:          clusterCmdOpts.port,
				Verbose:       rootCmdOpts.logVerbose,
				Debug:         rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			cluster, err := c.Init(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize k8s cluster: %w", err)
			}

			fmt.Printf("Bootstrapped k8s cluster on %q (%s).\n", cluster.Name, cluster.Address)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
}
