package k8sd

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/spf13/cobra"
)

var (
	sqlCmd = &cobra.Command{
		Use:    "sql <query>",
		Short:  "Execute an SQL query against the daemon",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return fmt.Errorf("invalid query")
			}
			cluster, err := app.New(cmd.Context(), app.Config{
				Debug:    rootCmdOpts.logDebug,
				Verbose:  rootCmdOpts.logVerbose,
				StateDir: rootCmdOpts.storageDir,
			})
			if err != nil {
				return fmt.Errorf("failed to create k8sd app: %w", err)
			}

			query := args[0]
			_, batch, err := cluster.MicroCluster.SQL(query)
			if err != nil {
				return fmt.Errorf("query failed: %w", err)
			}
			fmt.Println(batch.Results[0].Rows)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(sqlCmd)
}
