package k8s

import (
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/spf13/cobra"
)

func newXReconcileSnapdConfigCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "x-reconcile-snapd-config",
		Short:  "Reconcile k8s snapd configuration",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			mode, empty, err := snapdconfig.ParseMeta(cmd.Context(), env.Snap)
			if err != nil {
				if !empty {
					cmd.PrintErrf("Error: failed to parse meta configuration: %v\n", err)
					env.Exit(1)
					return
				}

				cmd.PrintErrf("Warning: failed to parse meta configuration: %v\n", err)
				cmd.PrintErrf("Warning: ignoring further errors to prevent infinite loop\n")
				if setErr := snapdconfig.SetMeta(cmd.Context(), env.Snap, snapdconfig.Meta{
					APIVersion: "1.30",
					Orb:        "none",
					Error:      err.Error(),
				}); setErr != nil {
					cmd.PrintErrf("Warning: failed to set meta configuration to safe defaults: %v\n", setErr)
				}
				env.Exit(0)
				return
			}

			if mode.Orb == "none" {
				cmd.PrintErrln("Warning: meta.orb is none, do not do anything")
			}

			mode.Orb = "snapd"
			if err := snapdconfig.SetMeta(cmd.Context(), env.Snap, mode); err != nil {
				cmd.PrintErrf("Error: failed to set snapd configur")
			}
		},
	}

	return cmd
}
