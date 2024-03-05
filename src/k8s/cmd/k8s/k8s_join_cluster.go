package k8s

import (
	"fmt"
	"os"

	apiv1 "github.com/canonical/k8s/api/v1"
	v1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/lxd/lxd/util"
	"github.com/spf13/cobra"
)

var (
	joinClusterCmdOpts struct {
		name    string
		address string
	}
	joinClusterCmdErrorMsgs = map[error]string{
		apiv1.ErrAlreadyBootstrapped: "A bootstrap node cannot join a cluster as it is already in a cluster. " +
			"Consider reinstalling the k8s snap and then join it.",
		apiv1.ErrInvalidJoinToken: "The provided join token is not valid. " +
			"Make sure that the name provided in `k8s get-join-token` matches the hostname of the " +
			"joining node or assign another name with the `--name` flag",
	}
)

type JoinClusterResult struct {
	Name string `json:"name" yaml:"name"`
}

func (b JoinClusterResult) String() string {
	return fmt.Sprintf("Cluster services have started on %q.\nPlease allow some time for initial Kubernetes node registration.\n", b.Name)
}

func newJoinClusterCmd() *cobra.Command {
	joinNodeCmd := &cobra.Command{
		Use:     "join-cluster <join-token>",
		Short:   "Join a cluster using the provided token",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments: provide only the join token that was generated with `sudo k8s get-join-token <node-name>`")
			}
			if len(args) < 1 {
				return fmt.Errorf("missing argument: provide the join token that was generated with `sudo k8s get-join-token <node-name>`")
			}

			defer errors.Transform(&err, joinClusterCmdErrorMsgs)

			joinToken := args[0]

			// Use hostname as default node name
			if joinClusterCmdOpts.name == "" {
				hostname, err := os.Hostname()
				if err != nil {
					return fmt.Errorf("--name is not set and failed to get hostname: %w", err)
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

			fmt.Fprintln(cmd.ErrOrStderr(), "Joining the cluster. This may take some time, please wait.")
			if err := k8sdClient.JoinCluster(cmd.Context(), joinClusterCmdOpts.name, joinClusterCmdOpts.address, joinToken); err != nil {
				return fmt.Errorf("failed to join cluster: %w", err)
			}

			f, err := formatter.New(rootCmdOpts.outputFormat, cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return f.Print(JoinClusterResult{
				Name: joinClusterCmdOpts.name,
			})
		},
	}
	joinNodeCmd.Flags().StringVar(&joinClusterCmdOpts.name, "name", "", "the name of the joining node. defaults to hostname")
	joinNodeCmd.Flags().StringVar(&joinClusterCmdOpts.address, "address", "", "the address (IP:Port) on which the nodes REST API should be available")
	return joinNodeCmd
}
