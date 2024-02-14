package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/spf13/cobra"
)

var (
	statusCmdOpts struct {
		outputFormat string
		timeout      time.Duration
		waitReady    bool
	}
)

func newStatusCmd() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:               "status",
		Short:             "Retrieve the current status of the cluster",
		Hidden:            true,
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			const minTimeout = 3 * time.Second
			if statusCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", statusCmdOpts.timeout, minTimeout, minTimeout)
				statusCmdOpts.timeout = minTimeout
			}

			timeoutCtx, cancel := context.WithTimeout(cmd.Context(), statusCmdOpts.timeout)
			defer cancel()
			clusterStatus, err := k8sdClient.ClusterStatus(timeoutCtx, statusCmdOpts.waitReady)
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
	statusCmd.PersistentFlags().StringVar(&statusCmdOpts.outputFormat, "format", "plain", "Specify in which format the output should be printed. One of plain, json or yaml")
	statusCmd.PersistentFlags().DurationVar(&statusCmdOpts.timeout, "timeout", 90*time.Second, "The max time to wait for the K8s API server to be ready.")
	statusCmd.PersistentFlags().BoolVar(&statusCmdOpts.waitReady, "wait-ready", false, "If set, the command will block until at least one cluster node is ready.")
	return statusCmd
}
