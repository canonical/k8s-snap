package k8sd

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/spf13/cobra"
)

var (
	sqlCmdOpts struct {
		verbose  bool
		debug    bool
		stateDir string
	}

	sqlCmd = &cobra.Command{
		Use:    "sql <query>",
		Short:  "Execute an SQL query against the daemon",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return fmt.Errorf("invalid query")
			}
			cluster, err := app.New(cmd.Context(), app.Config{
				Debug:    sqlCmdOpts.debug,
				Verbose:  sqlCmdOpts.verbose,
				StateDir: sqlCmdOpts.stateDir,
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
	sqlCmd.Flags().BoolVar(&sqlCmdOpts.debug, "debug", false, "")
	sqlCmd.Flags().BoolVar(&sqlCmdOpts.verbose, "verbose", false, "")
	sqlCmd.Flags().StringVar(&sqlCmdOpts.stateDir, "state-dir", "./build/stage", "")
	rootCmd.AddCommand(sqlCmd)
}
