package k8s

import (
	"fmt"

	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	statusCmdOpts struct {
		outputFormat string
	}

	statusCmd = &cobra.Command{
		Use:    "status",
		Short:  "Retrieve the current status of the cluster",
		Hidden: true,
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

			clusterStatus, err := c.ClusterStatus(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get cluster status: %w", err)
			}

			f, err := formatter.New(statusCmdOpts.outputFormat, cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return f.Print(clusterStatus)
		},
	}
)

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.PersistentFlags().StringVar(&statusCmdOpts.outputFormat, "format", "plain", "Specify in which format the output should be printed. One of plain, json or yaml")
}
