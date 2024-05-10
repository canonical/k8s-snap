package k8s

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"unicode"

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
		interactive  bool
		configFile   string
		name         string
		address      string
		outputFormat string
	}
	cmd := &cobra.Command{
		Use:    "bootstrap",
		Short:  "Bootstrap a new Kubernetes cluster",
		Long:   "Generate certificates, configure service arguments and start the Kubernetes services.",
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
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

			var bootstrapConfig apiv1.BootstrapConfig
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
				// Default bootstrap configuration
				bootstrapConfig = apiv1.BootstrapConfig{
					ClusterConfig: apiv1.UserFacingClusterConfig{
						Network: apiv1.NetworkConfig{
							Enabled: utils.Pointer(true),
						},
						DNS: apiv1.DNSConfig{
							Enabled: utils.Pointer(true),
						},
						Gateway: apiv1.GatewayConfig{
							Enabled: utils.Pointer(true),
						},
						LocalStorage: apiv1.LocalStorageConfig{
							Enabled: utils.Pointer(true),
						},
					},
				}
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

			outputFormatter.Print(BootstrapResult{Node: node})
		},
	}

	cmd.Flags().BoolVar(&opts.interactive, "interactive", false, "interactively configure the most important cluster options")
	cmd.Flags().StringVar(&opts.configFile, "file", "", "path to the YAML file containing your custom cluster bootstrap configuration.")
	cmd.Flags().StringVar(&opts.name, "name", "", "node name, defaults to hostname")
	cmd.Flags().StringVar(&opts.address, "address", "", "microcluster address, defaults to the node IP address")
	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")

	return cmd
}

func getConfigFromYaml(filePath string) (apiv1.BootstrapConfig, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return apiv1.BootstrapConfig{}, fmt.Errorf("failed to read file: %w", err)
	}

	var config apiv1.BootstrapConfig
	if err := yaml.UnmarshalStrict(b, &config); err != nil {
		return apiv1.BootstrapConfig{}, fmt.Errorf("failed to parse YAML config file: %w", err)
	}

	return config, nil
}

func getConfigInteractively(stdin io.Reader, stdout io.Writer, stderr io.Writer) apiv1.BootstrapConfig {
	config := apiv1.BootstrapConfig{}

	components := askQuestion(
		stdin, stdout, stderr,
		"Which features would you like to enable?",
		featureList,
		"network, dns, gateway, local-storage",
		nil,
	)
	for _, component := range strings.FieldsFunc(components, func(r rune) bool { return unicode.IsSpace(r) || r == ',' }) {
		switch component {
		case "network":
			config.ClusterConfig.Network.Enabled = utils.Pointer(true)
		case "dns":
			config.ClusterConfig.DNS.Enabled = utils.Pointer(true)
		case "ingress":
			config.ClusterConfig.Ingress.Enabled = utils.Pointer(true)
		case "load-balancer":
			config.ClusterConfig.LoadBalancer.Enabled = utils.Pointer(true)
		case "gateway":
			config.ClusterConfig.Gateway.Enabled = utils.Pointer(true)
		case "local-storage":
			config.ClusterConfig.LocalStorage.Enabled = utils.Pointer(true)
		}
	}

	podCIDR := askQuestion(stdin, stdout, stderr, "Please set the Pod CIDR:", nil, "10.1.0.0/16", nil)
	serviceCIDR := askQuestion(stdin, stdout, stderr, "Please set the Service CIDR:", nil, "10.152.183.0/24", nil)

	config.PodCIDR = utils.Pointer(podCIDR)
	config.ServiceCIDR = utils.Pointer(serviceCIDR)

	// TODO: any other configs we care about in the interactive bootstrap?

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
