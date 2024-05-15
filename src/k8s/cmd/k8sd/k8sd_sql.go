package k8sd

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/spf13/cobra"
)

func newSqlCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	return &cobra.Command{
		Use:    "sql <query>",
		Short:  "Execute an SQL query against the daemon",
		Hidden: true,
		Args:   cmdutil.ExactArgs(env, 1),
		Run: func(cmd *cobra.Command, args []string) {
			app, err := app.New(app.Config{
				StateDir: rootCmdOpts.stateDir,
				Snap:     env.Snap,
			})
			if err != nil {
				cmd.PrintErrf("Error: Failed to initialize k8sd app.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			str, batch, err := app.MicroCluster().SQL(cmd.Context(), args[0])
			if err != nil {
				cmd.PrintErrf("Error: Failed to execute the SQL query.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if args[0] == ".dump" || args[0] == ".schema" {
				cmd.Println(str)
			} else {
				cmd.Println(batch.Results[0].Rows)
			}
		},
	}
}
