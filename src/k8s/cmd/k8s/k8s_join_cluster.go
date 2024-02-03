package k8s

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/lxd/lxd/util"
	"github.com/spf13/cobra"
)

var (
	joinNodeCmdOpts struct {
		name    string
		address string
		timeout time.Duration
	}

	joinNodeCmd = &cobra.Command{
		Use:   "join-cluster <token>",
		Short: "Join a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token := args[0]

			// Use hostname as default node name
			if joinNodeCmdOpts.name == "" {
				hostname, err := os.Hostname()
				if err != nil {
					return fmt.Errorf("--name is not set and failed to get hostname: %w", err)
				}
				joinNodeCmdOpts.name = hostname
			}

			if joinNodeCmdOpts.address == "" {
				joinNodeCmdOpts.address = util.CanonicalNetworkAddress(
					util.NetworkInterfaceAddress(), config.DefaultPort,
				)
			}

			c, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				StateDir: clusterCmdOpts.stateDir,
				Verbose:  rootCmdOpts.logVerbose,
				Debug:    rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create cluster client: %w", err)
			}

			if c.IsBootstrapped(cmd.Context()) {
				return fmt.Errorf("k8s cluster already bootstrapped")
			}
			const minTimeout = 3 * time.Second
			if joinNodeCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v, using the minimum %v instead.\n", joinNodeCmdOpts.timeout, minTimeout, minTimeout)
				joinNodeCmdOpts.timeout = minTimeout
			}

			timeoutCtx, cancel := context.WithTimeout(cmd.Context(), joinNodeCmdOpts.timeout)
			defer cancel()

			err = c.JoinNode(timeoutCtx, joinNodeCmdOpts.name, joinNodeCmdOpts.address, token)
			if err != nil {
				return fmt.Errorf("failed to join cluster: %w", err)
			}

			fmt.Println("Joined the cluster.")
			return nil
		},
	}
)

func init() {
	joinNodeCmd.Flags().StringVar(&joinNodeCmdOpts.name, "name", "", "The name of the joining node. defaults to hostname")
	joinNodeCmd.Flags().StringVar(&joinNodeCmdOpts.address, "address", "", "The address (IP:Port) on which the nodes REST API should be available")
	joinNodeCmd.Flags().DurationVar(&joinNodeCmdOpts.timeout, "timeout", 90*time.Second, "The max time to wait for the node to be ready.")

	rootCmd.AddCommand(joinNodeCmd)
}
