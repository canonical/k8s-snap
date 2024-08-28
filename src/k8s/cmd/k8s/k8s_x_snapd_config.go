package k8s

import (
	"context"
	"time"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/spf13/cobra"
)

func newXSnapdConfigCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		timeout time.Duration
	}

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

			if mode.Orb == "none" {
				cmd.PrintErrln("Warning: meta.orb is none, skipping reconcile actions")
				return
			}

			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: failed to create k8sd client: %v\n", err)
				env.Exit(1)
				return
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			defer cancel()
			if err := control.WaitUntilReady(ctx, func() (bool, error) {
				_, partOfCluster, err := client.NodeStatus(cmd.Context())
				if !partOfCluster {
					cmd.PrintErrf("Node is not part of a cluster: %v\n", err.Error())
					env.Exit(1)
				}
				return err == nil, nil
			}); err != nil {
				cmd.PrintErrf("Error: Node did not come up in time: %v\n", err)
				env.Exit(1)
			}

			switch mode.Orb {
			case "k8sd":
				response, err := client.GetClusterConfig(cmd.Context())
				if err != nil {
					cmd.PrintErrf("Error: failed to retrieve cluster configuration: %v\n", err)
					env.Exit(1)
					return
				}
				if err := snapdconfig.SetSnapdFromK8sd(cmd.Context(), response.Config, env.Snap); err != nil {
					cmd.PrintErrf("Error: failed to update snapd state: %v\n", err)
					env.Exit(1)
					return
				}
			case "snapd":
				if err := snapdconfig.SetK8sdFromSnapd(cmd.Context(), client, env.Snap); err != nil {
					cmd.PrintErrf("Error: failed to update k8sd state: %v\n", err)
					env.Exit(1)
					return
				}
			}

			mode.Orb = "snapd"
			if err := snapdconfig.SetMeta(cmd.Context(), env.Snap, mode); err != nil {
				cmd.PrintErrf("Error: failed to set snapd configuration: %v\n", err)
				env.Exit(1)
				return
			}
		},
	}
	reconcileCmd.Flags().DurationVar(&opts.timeout, "timeout", 1*time.Minute, "maximum time to wait")

	cmd := &cobra.Command{
		Use:    "x-snapd-config",
		Short:  "Manage snapd configuration",
		Hidden: true,
	}

	cmd.AddCommand(reconcileCmd)
	cmd.AddCommand(disableCmd)

	return cmd
}
