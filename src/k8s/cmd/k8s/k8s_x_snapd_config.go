package k8s

import (
	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/spf13/cobra"
)

func newXSnapdConfigCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	disableCmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable the use of snap get/set to manage the cluster configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := snapdconfig.Disable(cmd.Context(), env.Snap); err != nil {
				cmd.PrintErrf("Error: failed to disable snapd configuration: %v\n", err)
				env.Exit(1)
			}
		},
	}
	reconcileCmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile the cluster configuration changes from k8s {set,get} <-> snap {set,get} k8s",
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
				return
			}

			switch mode.Orb {
			case "none":
				cmd.PrintErrln("Warning: meta.orb is none, skipping reconcile actions")
				return
			case "k8sd":
				client, err := env.Client(cmd.Context())
				if err != nil {
					cmd.PrintErrf("Error: failed to create k8sd client: %v\n", err)
					env.Exit(1)
					return
				}
				config, err := client.GetClusterConfig(cmd.Context(), apiv1.GetClusterConfigRequest{})
				if err != nil {
					cmd.PrintErrf("Error: failed to retrieve cluster configuration: %v\n", err)
					env.Exit(1)
					return
				}
				if err := snapdconfig.SetSnapdFromK8sd(cmd.Context(), config, env.Snap); err != nil {
					cmd.PrintErrf("Error: failed to update snapd state: %v\n", err)
					env.Exit(1)
					return
				}
			case "snapd":
				client, err := env.Client(cmd.Context())
				if err != nil {
					cmd.PrintErrf("Error: failed to create k8sd client: %v\n", err)
					env.Exit(1)
					return
				}
				if err := snapdconfig.SetK8sdFromSnapd(cmd.Context(), client, env.Snap); err != nil {
					cmd.PrintErrf("Error: failed to update k8sd state: %v\n", err)
					env.Exit(1)
					return
				}
			}

			mode.Orb = "k8sd"
			if err := snapdconfig.SetMeta(cmd.Context(), env.Snap, mode); err != nil {
				cmd.PrintErrf("Error: failed to set snapd configuration: %v\n", err)
				env.Exit(1)
				return
			}
		},
	}
	cmd := &cobra.Command{
		Use:    "x-snapd-config",
		Short:  "Manage snapd configuration",
		Hidden: true,
	}

	cmd.AddCommand(reconcileCmd)
	cmd.AddCommand(disableCmd)

	return cmd
}
