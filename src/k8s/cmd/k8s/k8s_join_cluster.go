package k8s

import (
	"fmt"
	"io"
	"os"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/utils"
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
		name         string
		address      string
		configFile   string
		outputFormat string
		timeout      time.Duration
	}
	cmd := &cobra.Command{
		Use:    "join-cluster <join-token>",
		Short:  "Join a cluster using the provided token",
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Args:   cmdutil.ExactArgs(env, 1),
		Run: func(cmd *cobra.Command, args []string) {
			token := args[0]

			if opts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", opts.timeout, minTimeout, minTimeout)
				opts.timeout = minTimeout
			}

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

			address, err := utils.ParseAddressString(opts.address, config.DefaultPort)
			if err != nil {
				cmd.PrintErrf("Error: Failed to parse the address %q.\n\nThe error was: %v\n", opts.address, err)
				env.Exit(1)
				return
			}

			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if _, initialized, err := client.NodeStatus(cmd.Context()); err != nil {
				cmd.PrintErrf("Error: Failed to check the current node status.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			} else if initialized {
				cmd.PrintErrln("Error: The node is already part of a cluster")
				env.Exit(1)
				return
			}

			var joinClusterConfig string
			if opts.configFile != "" {
				var b []byte
				var err error

				if opts.configFile == "-" {
					b, err = io.ReadAll(os.Stdin)
					if err != nil {
						cmd.PrintErrf("Error: Failed to read join configuration from stdin. \n\nThe error was: %v\n", err)
						env.Exit(1)
						return
					}
				} else {
					b, err = os.ReadFile(opts.configFile)
					if err != nil {
						cmd.PrintErrf("Error: Failed to read join configuration from %q.\n\nThe error was: %v\n", opts.configFile, err)
						env.Exit(1)
						return
					}
				}
				joinClusterConfig = string(b)
			}

			cmd.PrintErrln("Joining the cluster. This may take a few seconds, please wait.")
			if err := client.JoinCluster(cmd.Context(), apiv1.JoinClusterRequest{
				Name:    opts.name,
				Address: address,
				Token:   token,
				Config:  joinClusterConfig,
				Timeout: opts.timeout,
			}); err != nil {
				cmd.PrintErrf("Error: Failed to join the cluster using the provided token.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			outputFormatter.Print(JoinClusterResult{Name: opts.name})
		},
	}
	cmd.Flags().StringVar(&opts.name, "name", "", "node name, defaults to hostname")
	cmd.Flags().StringVar(&opts.address, "address", "", "microcluster address or CIDR, defaults to the node IP address")
	cmd.Flags().StringVar(&opts.configFile, "file", "", "path to the YAML file containing your custom cluster join configuration. Use '-' to read from stdin.")
	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 90*time.Second, "the max time to wait for the command to execute")
	return cmd
}
