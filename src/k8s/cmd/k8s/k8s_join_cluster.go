package k8s

import (
	"fmt"
	"os"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/lxd/lxd/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type JoinClusterResult struct {
	Name string `json:"name" yaml:"name"`
}

func (b JoinClusterResult) String() string {
	return fmt.Sprintf("Cluster services have started on %q.\nPlease allow some time for initial Kubernetes node registration.\n", b.Name)
}

func newJoinClusterCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		name       string
		address    string
		configFile string
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
				opts.address = util.NetworkInterfaceAddress()
			}
			opts.address = util.CanonicalNetworkAddress(opts.address, config.DefaultPort)

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

			joinClusterConfig := apiv1.JoinClusterConfig{}
			if opts.configFile != "" {
				joinClusterConfig, err = getJoinClusterConfigFromYaml(opts.configFile)
				if err != nil {
					cmd.PrintErrf("Error: Failed to read join configuration from %q.\n\nThe error was: %v\n", opts.configFile, err)
					env.Exit(1)
					return
				}
			}

			cmd.PrintErrln("Joining the cluster. This may take a few seconds, please wait.")
			if err := client.JoinCluster(cmd.Context(), apiv1.JoinClusterRequest{Name: opts.name, Address: opts.address, Token: token, Config: joinClusterConfig}); err != nil {
				cmd.PrintErrf("Error: Failed to join the cluster using the provided token.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(JoinClusterResult{Name: opts.name}); err != nil {
				cmd.PrintErrf("WARNING: Failed to print the join cluster result.\n\nThe error was: %v\n", err)
			}
		},
	}
	cmd.Flags().StringVar(&opts.name, "name", "", "node name, defaults to hostname")
	cmd.Flags().StringVar(&opts.address, "address", "", "microcluster address, defaults to the node IP address")
	cmd.PersistentFlags().StringVar(&opts.configFile, "config", "", "path to the YAML file containing your custom cluster join configuration")
	return cmd
}

func getJoinClusterConfigFromYaml(filePath string) (apiv1.JoinClusterConfig, error) {
	config := apiv1.JoinClusterConfig{}

	yamlContent, err := os.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("failed to read YAML config file: %w", err)
	}

	err = yaml.Unmarshal(yamlContent, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse YAML config file: %w", err)
	}

	return config, nil
}
