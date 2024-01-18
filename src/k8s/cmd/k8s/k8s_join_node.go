package k8s

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/canonical/lxd/lxd/util"
	"github.com/spf13/cobra"
)

var (
	joinNodeCmdOpts struct {
		name    string
		address string
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

			client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				StorageDir: clusterCmdOpts.storageDir,
				Verbose:    rootCmdOpts.logVerbose,
				Debug:      rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create cluster client: %w", err)
			}

			err = client.JoinNode(cmd.Context(), joinNodeCmdOpts.name, joinNodeCmdOpts.address, token)
			if err != nil {
				return fmt.Errorf("failed to join cluster: %w", err)
			}

			fmt.Printf("Joined %s (%s) to cluster.\n", joinNodeCmdOpts.name, joinNodeCmdOpts.address)
			return nil
		},
	}
)

func init() {
	joinNodeCmd.Flags().StringVar(&joinNodeCmdOpts.name, "name", "", "The name of the joining node. defaults to hostname")
	joinNodeCmd.Flags().StringVar(&joinNodeCmdOpts.address, "address", "", "The address (IP:Port) on which the nodes REST API should be available")

	rootCmd.AddCommand(joinNodeCmd)
}
