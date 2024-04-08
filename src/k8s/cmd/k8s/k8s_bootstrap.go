package k8s

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type BootstrapResult struct {
	Node apiv1.NodeStatus `json:"node" yaml:"node"`
}

func (b BootstrapResult) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("Bootstrapped a new Kubernetes cluster with node address %q.\n", b.Node.Address))
	buf.WriteString("The node will be 'Ready' to host workloads after the CNI is deployed successfully.\n")

	return buf.String()
}

func newBootstrapCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		interactive bool
		configFile  string
		name        string
		address     string
	}
	cmd := &cobra.Command{
		Use:    "bootstrap",
		Short:  "Bootstrap a new Kubernetes cluster",
		Long:   "Generate certificates, configure service arguments and start the Kubernetes services.",
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			if opts.interactive && opts.configFile != "" {
				cmd.PrintErrln("Error: --interactive and --file flags cannot be set at the same time.")
				env.Exit(1)
				return
			}

			// Use hostname as default node name
			if opts.name == "" {
				hostname, err := os.Hostname()
				if err != nil {
					cmd.PrintErrf("Error: --name is not set and could not determine the current node name.\n\nThe error was: %v\n", err)
					env.Exit(1)
					return
				}
				opts.name, err = utils.CleanHostname(hostname)
				if err != nil {
					cmd.PrintErrf("Error: --name is not set and default hostname %q is not valid.\n\nThe error was: %v\n", hostname, err)
				}
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

			bootstrapConfig := apiv1.BootstrapConfig{}
			switch {
			case opts.interactive:
				bootstrapConfig = getConfigInteractively(env.Stdin, env.Stdout, env.Stderr)
			case opts.configFile != "":
				bootstrapConfig, err = getConfigFromYaml(opts.configFile)
				if err != nil {
					cmd.PrintErrf("Error: Failed to read bootstrap configuration from %q.\n\nThe error was: %v\n", opts.configFile, err)
					env.Exit(1)
					return
				}
			default:
				bootstrapConfig.SetDefaults()
			}

			cmd.PrintErrln("Bootstrapping the cluster. This may take a few seconds, please wait.")

			request := apiv1.PostClusterBootstrapRequest{
				Name:    opts.name,
				Address: opts.address,
				Config:  bootstrapConfig,
			}

			node, err := client.Bootstrap(cmd.Context(), request)
			if err != nil {
				cmd.PrintErrf("Error: Failed to bootstrap the cluster.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(BootstrapResult{Node: node}); err != nil {
				cmd.PrintErrf("WARNING: Failed to print the cluster bootstrap result.\n\nThe error was: %v\n", err)
			}
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.interactive, "interactive", false, "interactively configure the most important cluster options")
	cmd.PersistentFlags().StringVar(&opts.configFile, "file", "", "path to the YAML file containing your custom cluster bootstrap configuration.")
	cmd.Flags().StringVar(&opts.name, "name", "", "node name, defaults to hostname")
	cmd.Flags().StringVar(&opts.address, "address", "", "microcluster address, defaults to the node IP address")

	return cmd
}

func getConfigFromYaml(filePath string) (apiv1.BootstrapConfig, error) {
	config := apiv1.BootstrapConfig{}
	config.SetDefaults()

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

func getConfigInteractively(stdin io.Reader, stdout io.Writer, stderr io.Writer) apiv1.BootstrapConfig {
	config := apiv1.BootstrapConfig{}
	config.SetDefaults()

	components := askQuestion(
		stdin, stdout, stderr,
		"Which components would you like to enable?",
		componentList,
		strings.Join(config.Components, ", "),
		nil,
	)
	config.Components = strings.Split(components, ",")

	config.ClusterCIDR = askQuestion(stdin, stdout, stderr, "Please set the Cluster CIDR:", nil, config.ClusterCIDR, nil)
	config.ServiceCIDR = askQuestion(stdin, stdout, stderr, "Please set the Service CIDR:", nil, config.ServiceCIDR, nil)
	rbac := askBool(stdin, stdout, stderr, "Enable Role Based Access Control (RBAC)?", []string{"yes", "no"}, "yes")
	*config.EnableRBAC = rbac
	return config
}

// askQuestion will ask the user for input.
// askQuestion will keep asking if the input is not valid.
// askQuestion will remove all whitespaces and capitalization of the input.
// customErr can be used to provide extra error messages for specific non-valid inputs.
func askQuestion(stdin io.Reader, stdout io.Writer, stderr io.Writer, question string, options []string, defaultVal string, customErr map[string]string) string {
	for {
		q := question
		if options != nil {
			q = fmt.Sprintf("%s (%s)", q, strings.Join(options, ", "))
		}
		if defaultVal != "" {
			q = fmt.Sprintf("%s [%s]:", q, defaultVal)
		}
		q = fmt.Sprintf("%s ", q)

		var s string
		r := bufio.NewReader(stdin)
		for {
			fmt.Fprint(stdout, q)
			s, _ = r.ReadString('\n')
			if s != "" {
				break
			}
		}
		s = strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(s), " ", ""), "\n", "")

		if s == "" {
			return defaultVal
		}

		// Check if the input is valid
		if options != nil || len(options) > 0 {
			valid := true
			sSlice := strings.Split(s, ",")

			for _, element := range sSlice {
				if !slices.Contains(options, element) {
					if msg, ok := customErr[element]; ok {
						fmt.Fprintf(stderr, "  %s\n", msg)
					} else {
						fmt.Fprintf(stderr, "  %q is not a valid option.\n", element)
					}
					valid = false
				}
			}
			if !valid {
				continue
			}
		}
		return s
	}
}

// askBool asks a question and expect a yes/no answer.
func askBool(stdin io.Reader, stdout io.Writer, stderr io.Writer, question string, options []string, defaultVal string) bool {
	for {
		answer := askQuestion(stdin, stdout, stderr, question, options, defaultVal, nil)

		if utils.ValueInSlice(strings.ToLower(answer), []string{"yes", "y"}) {
			return true
		} else if utils.ValueInSlice(strings.ToLower(answer), []string{"no", "n"}) {
			return false
		}

		fmt.Fprintf(stderr, "Invalid input, try again.\n\n")
	}
}
