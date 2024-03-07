package k8s

import (
	"fmt"
	"os"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/lxd/lxd/util"
	"github.com/spf13/cobra"
)

var (
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

func newJoinClusterCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		name    string
		address string
	}
	cmd := &cobra.Command{
		Use:    "join-cluster <join-token>",
		Short:  "Join a cluster using the provided token",
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Args:   cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			token := args[0]

			// Use hostname as default node name
			if opts.name == "" {
				// TODO(neoaggelos): use the encoded node name from the token, if available.
				hostname, err := os.Hostname()
				if err != nil {
					cmd.PrintErrf("ERROR: --name is not set and could not determine the current node name.\n\nThe error was: %v\n", err)
					env.Exit(1)
					return
				}
				opts.name = hostname
			}

			if opts.address == "" {
				opts.address = util.CanonicalNetworkAddress(util.NetworkInterfaceAddress(), config.DefaultPort)
			}

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("ERROR: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if client.IsBootstrapped(cmd.Context()) {
				cmd.PrintErrln("ERROR: The node is already part of a cluster")
				env.Exit(1)
				return
			}

			cmd.PrintErrln("Joining the cluster. This may take a few seconds, please wait.")
			if err := client.JoinCluster(cmd.Context(), opts.name, opts.address, token); err != nil {
				cmd.PrintErrln("ERROR: Failed to join the cluster using the provided token.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(JoinClusterResult{Name: opts.name}); err != nil {
				cmd.PrintErrf("WARNING: Failed to print the join cluster result.\n\nThe error was: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVar(&opts.name, "name", "", "the name of the joining node. defaults to hostname")
	cmd.Flags().StringVar(&opts.address, "address", "", "the address (IP:Port) on which the nodes REST API should be available")
	return cmd
}
