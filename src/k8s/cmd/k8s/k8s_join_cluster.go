package k8s

import (
	"context"
	"fmt"
	"os"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/utils"
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

			var joinClusterConfig string
			if opts.configFile != "" {
				joinClusterConfig, err = readAndParseConfigFile(env, opts.configFile)
				if err != nil {
					cmd.PrintErrf("Error: Failed to read config file %s.\n\n The error was %v\n", opts.configFile, err)
					env.Exit(1)
					return
				}
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), opts.timeout)
			cobra.OnFinalize(cancel)

			cmd.PrintErrln("Joining the cluster. This may take a few seconds, please wait.")
			if err := client.JoinCluster(ctx, apiv1.JoinClusterRequest{Name: opts.name, Address: address, Token: token, Config: joinClusterConfig}); err != nil {
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

// readAndParseConfigFile reads the join configuration file and returns the parsed configuration as a string.
// readAndParseConfigFile replaces file paths in "extra-node-config-files" with its contents.
// TODO(bschimke): We don't use explicit types for the join configs because control plane and worker join configs are different.
// This leads to the ugly type assertion dance below. We should consider generics for that instead.
func readAndParseConfigFile(env cmdutil.ExecutionEnvironment, configFilePath string) (string, error) {
	joinClusterConfigMap, err := getConfigFromYaml[map[string]any](env, configFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read and parse the join configuration file: %w", err)
	}

	if ef, ok := joinClusterConfigMap["extra-node-config-files"]; ok {
		if extraFiles, ok := ef.([]interface{}); ok {
			// Resolve file names to file contents before sending to the server.
			for idx, configFile := range extraFiles {
				content, err := os.ReadFile(configFile.(string))
				if err != nil {
					return "", fmt.Errorf("failed to read extra node config file %q: %w", configFile, err)
				}
				extraFiles[idx] = string(content)
			}
			joinClusterConfigMap["extra-node-config-files"] = extraFiles
		}
	}

	b, err := yaml.Marshal(joinClusterConfigMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal the join configuration: %w", err)
	}
	return string(b), nil
}
