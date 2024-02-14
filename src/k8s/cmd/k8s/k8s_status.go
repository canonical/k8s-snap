package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var (
	statusCmdOpts struct {
		outputFormat string
		timeout      time.Duration
		waitReady    bool
	}

	statusCmd = &cobra.Command{
		Use:    "status",
		Short:  "Retrieve the current status of the cluster",
		Hidden: true,
		Run: ew.Run(func(cmd *cobra.Command, args []string) error {
			c, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				StateDir: clusterCmdOpts.stateDir,
				Verbose:  rootCmdOpts.logVerbose,
				Debug:    rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			const minTimeout = 3 * time.Second
			if statusCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", statusCmdOpts.timeout, minTimeout, minTimeout)
				statusCmdOpts.timeout = minTimeout
			}

			timeoutCtx, cancel := context.WithTimeout(cmd.Context(), statusCmdOpts.timeout)
			defer cancel()
			clusterStatus, err := c.ClusterStatus(timeoutCtx, statusCmdOpts.waitReady)
			if err != nil {
				return fmt.Errorf("failed to get cluster status: %w", err)
			}

			f, err := formatter.New(statusCmdOpts.outputFormat, cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return f.Print(clusterStatus)
		}),
	}
)

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.PersistentFlags().StringVar(&statusCmdOpts.outputFormat, "format", "plain", "Specify in which format the output should be printed. One of plain, json or yaml")
	statusCmd.PersistentFlags().DurationVar(&statusCmdOpts.timeout, "timeout", 90*time.Second, "The max time to wait for the K8s API server to be ready.")
	statusCmd.PersistentFlags().BoolVar(&statusCmdOpts.waitReady, "wait-ready", false, "If set, the command will block until at least one cluster node is ready.")
}
