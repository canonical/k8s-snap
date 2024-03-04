package k8s

import (
	"context"
	"fmt"
	"os"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	v1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/lxd/lxd/util"
	"github.com/spf13/cobra"
)

var (
	joinClusterCmdOpts struct {
		name    string
		address string
		timeout time.Duration
	}
	joinClusterCmdErrorMsgs = map[error]string{
		apiv1.ErrUnknown: "An error occurred while joining the cluster:\n",
		apiv1.ErrAlreadyBootstrapped: "A bootstrap node cannot join a cluster as it is already in a cluster. " +
			"Consider reinstalling the k8s snap and then join it.",
		apiv1.ErrInvalidJoinToken: "The provided join token is not valid. " +
			"Make sure that the name provided in `k8s get-join-token` matches the hostname of the " +
			"joining node or assign another name with the `--name` flag",
	}
)

func newJoinClusterCmd() *cobra.Command {
	joinNodeCmd := &cobra.Command{
		Use:     "join-cluster <join-token>",
		Short:   "Join a cluster",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) > 1 {
				return fmt.Errorf("Too many arguments. Please provide only the join token that was generated with `sudo k8s get-join-token <node-name>`")
			}
			if len(args) < 1 {
				return fmt.Errorf("Please provide the join token that was generated with `sudo k8s get-join-token <node-name>`")
			}

			defer errors.Transform(&err, joinClusterCmdErrorMsgs)

			joinToken := args[0]

			// Use hostname as default node name
			if joinClusterCmdOpts.name == "" {
				hostname, err := os.Hostname()
				if err != nil {
					return fmt.Errorf("Could not determine the node's hostname and '--name' is not set: %w", err)
				}
				joinClusterCmdOpts.name = hostname
			}

			if joinClusterCmdOpts.address == "" {
				joinClusterCmdOpts.address = util.CanonicalNetworkAddress(
					util.NetworkInterfaceAddress(), config.DefaultPort,
				)
			}

			if k8sdClient.IsBootstrapped(cmd.Context()) {
				return v1.ErrAlreadyBootstrapped
			}
			const minTimeout = 3 * time.Second
			if joinClusterCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v, using the minimum %v instead.\n", joinClusterCmdOpts.timeout, minTimeout, minTimeout)
				joinClusterCmdOpts.timeout = minTimeout
			}

			timeoutCtx, cancel := context.WithTimeout(cmd.Context(), joinClusterCmdOpts.timeout)
			defer cancel()

			fmt.Println("Joining the cluster. This may take some time, please wait.")
			if err := k8sdClient.JoinCluster(timeoutCtx, joinClusterCmdOpts.name, joinClusterCmdOpts.address, joinToken); err != nil {
				return fmt.Errorf("Failed to join cluster: %w", err)
			}

			fmt.Printf("Joined the cluster as %q.\nPlease allow some time for Kubernetes node registration.\n", joinClusterCmdOpts.name)
			return nil
		},
	}
	joinNodeCmd.Flags().StringVar(&joinClusterCmdOpts.name, "name", "", "The name of the joining node. defaults to hostname")
	joinNodeCmd.Flags().StringVar(&joinClusterCmdOpts.address, "address", "", "The address (IP:Port) on which the nodes REST API should be available")
	joinNodeCmd.Flags().DurationVar(&joinClusterCmdOpts.timeout, "timeout", 90*time.Second, "The max time to wait for the node to be ready.")
	return joinNodeCmd
}
