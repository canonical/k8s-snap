package k8s

import (
	"fmt"
	"os"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/lxd/lxd/util"
	"github.com/spf13/cobra"
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
		Args:   cmdutil.ExactArgs(env, 1),
		Run: func(cmd *cobra.Command, args []string) {
			token := args[0]

			// Use hostname as default node name
			if opts.name == "" {
				// TODO(neoaggelos): use the encoded node name from the token, if available.
				hostname, err := os.Hostname()
				if err != nil {
					cmd.PrintErrf("Error: --name is not set and could not determine the current node name.\n\nThe error was: %v\n", err)
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
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if client.IsBootstrapped(cmd.Context()) {
				cmd.PrintErrln("Error: The node is already part of a cluster")
				env.Exit(1)
				return
			}

			cmd.PrintErrln("Joining the cluster. This may take a few seconds, please wait.")
			if err := client.JoinCluster(cmd.Context(), opts.name, opts.address, token); err != nil {
				cmd.PrintErrf("Error: Failed to join the cluster using the provided token.\n\nThe error was: %v\n", err)
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
