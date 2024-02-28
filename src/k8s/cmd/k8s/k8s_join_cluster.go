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
	joinNodeCmdOpts struct {
		name    string
		address string
		timeout time.Duration
	}
	joinNodeCmdErrorMsgs = map[error]string{
		apiv1.ErrAlreadyBootstrapped: "A bootstrap node cannot join a cluster as it is already in a cluster. " +
			"Consider reinstalling the k8s snap and then join it.",
		apiv1.ErrInvalidJoinToken: "The provided token is not valid. " +
			"Make sure that the name provided in `k8s get-join-token` matches the hostname of the " +
			"joining node or asign another name with the `--name` flag",
	}
)

func newJoinNodeCmd() *cobra.Command {
	joinNodeCmd := &cobra.Command{
		Use:               "join-cluster <token>",
		Short:             "Join a cluster",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments: provide only the token that was generated with `sudo k8s get-join-token <node-name>`")
			}
			if len(args) < 1 {
				return fmt.Errorf("missing argument: provide the token that was generated with `sudo k8s get-join-token <node-name>`")
			}

			defer errors.Transform(&err, joinNodeCmdErrorMsgs)

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

			if k8sdClient.IsBootstrapped(cmd.Context()) {
				return v1.ErrAlreadyBootstrapped
			}
			const minTimeout = 3 * time.Second
			if joinNodeCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v, using the minimum %v instead.\n", joinNodeCmdOpts.timeout, minTimeout, minTimeout)
				joinNodeCmdOpts.timeout = minTimeout
			}

			timeoutCtx, cancel := context.WithTimeout(cmd.Context(), joinNodeCmdOpts.timeout)
			defer cancel()

			fmt.Println("Joining the cluster. This may take some time, please wait.")
			if err := k8sdClient.JoinCluster(timeoutCtx, joinNodeCmdOpts.name, joinNodeCmdOpts.address, token); err != nil {
				return fmt.Errorf("failed to join cluster: %w", err)
			}

			fmt.Printf("Joined the cluster as %q.\nPlease allow some time for Kubernetes node registration.\n", joinNodeCmdOpts.name)
			return nil
		},
	}
	joinNodeCmd.Flags().StringVar(&joinNodeCmdOpts.name, "name", "", "The name of the joining node. defaults to hostname")
	joinNodeCmd.Flags().StringVar(&joinNodeCmdOpts.address, "address", "", "The address (IP:Port) on which the nodes REST API should be available")
	joinNodeCmd.Flags().DurationVar(&joinNodeCmdOpts.timeout, "timeout", 90*time.Second, "The max time to wait for the node to be ready.")
	return joinNodeCmd
}
