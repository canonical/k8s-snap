package k8s

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	bootstrapCmdOpts struct {
		interactive bool
		config      string // yaml config filename
		timeout     time.Duration
	}

	bootstrapCmdErrorMsgs = map[error]string{
		apiv1.ErrUnknown:             "An unknown error occured while bootstrapping the cluster:\n",
		apiv1.ErrAlreadyBootstrapped: "K8s cluster already bootstrapped.",
	}
	bootstrappableComponents = []string{"network", "dns", "gateway", "ingress", "storage", "metrics-server"}
)

func newBootstrapCmd() *cobra.Command {
	bootstrapCmd := &cobra.Command{
		Use:     "bootstrap",
		Short:   "Bootstrap a k8s cluster on this node.",
		Long:    "Initialize the necessary folders, permissions, service arguments, certificates and start up the Kubernetes services.",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, bootstrapCmdErrorMsgs)

			if k8sdClient.IsBootstrapped(cmd.Context()) {
				return apiv1.ErrAlreadyBootstrapped
			}

			const minTimeout = 3 * time.Second
			if bootstrapCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", bootstrapCmdOpts.timeout, minTimeout, minTimeout)
				bootstrapCmdOpts.timeout = minTimeout
			}

			bootstrapConfig := apiv1.BootstrapConfig{}
			if bootstrapCmdOpts.interactive && bootstrapCmdOpts.config != "" {
				return fmt.Errorf("failed to bootstrap cluster: cannot use both --interactive and --config flags at the same time")
			}

			if bootstrapCmdOpts.interactive {
				bootstrapConfig = getConfigInteractively(cmd.Context())
			} else if bootstrapCmdOpts.config != "" {
				bootstrapConfig, err = getConfigFromYaml(bootstrapCmdOpts.config)
				if err != nil {
					return fmt.Errorf("failed to bootstrap cluster: %w", err)
				}
			} else {
				bootstrapConfig.SetDefaults()
			}

			fmt.Println("Bootstrapping the cluster. This may take some time, please wait.")
			cluster, err := k8sdClient.Bootstrap(cmd.Context(), bootstrapConfig)
			if err != nil {
				return fmt.Errorf("failed to bootstrap cluster: %w", err)
			}

			fmt.Printf("Cluster services have started on %q.\nPlease allow some time for initial Kubernetes node registration.\n", cluster.Name)
			return nil
		},
	}

	bootstrapCmd.PersistentFlags().BoolVar(&bootstrapCmdOpts.interactive, "interactive", false, "Interactively configure the most important cluster options.")
	bootstrapCmd.PersistentFlags().DurationVar(&bootstrapCmdOpts.timeout, "timeout", 90*time.Second, "The max time to wait for k8s to bootstrap.")

	return bootstrapCmd
}

func getConfigFromYaml(filePath string) (apiv1.BootstrapConfig, error) {
	config := apiv1.BootstrapConfig{}
	config.SetDefaults()

	// Read the yaml file
	yamlContent, err := os.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("failed to read YAML config file: %w", err)
	}
	// Parse the yaml file
	err = yaml.Unmarshal(yamlContent, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse YAML config file: %w", err)
	}

	return config, nil
}

func getConfigInteractively(ctx context.Context) apiv1.BootstrapConfig {
	config := apiv1.BootstrapConfig{}
	config.SetDefaults()

	components := askQuestion(
		"Which components would you like to enable?",
		bootstrappableComponents,
		strings.Join(config.Components, ", "),
		map[string]string{"loadbalancer": "The \"loadbalancer\" component requires manual configuration and needs to be enabled after bootstrapping the cluster."},
	)
	config.Components = strings.Split(components, ",")

	config.ClusterCIDR = askQuestion("Please set the Cluster CIDR:", nil, config.ClusterCIDR, nil)

	rbac := askBool("Enable Role Based Access Control (RBAC)?", []string{"yes", "no"}, "yes")
	*config.EnableRBAC = rbac
	return config
}

// askQuestion will ask the user for input.
// askQuestion will keep asking if the input is not valid.
// askQuestion will remove all whitespaces and capitalization of the input.
// customErr can be used to provide extra error messages for specific non-valid inputs.
func askQuestion(question string, options []string, defaultVal string, customErr map[string]string) string {
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
		r := bufio.NewReader(os.Stdin)
		for {
			fmt.Fprint(os.Stdout, q)
			s, _ = r.ReadString('\n')
			if s != "" {
				break
			}
		}
		s = strings.ReplaceAll(strings.ToLower(s), " ", "")

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
						fmt.Fprintf(os.Stdout, "  %s\n", msg)
					} else {
						fmt.Fprintf(os.Stdout, "  %q is not a valid option.\n", element)
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
func askBool(question string, options []string, defaultVal string) bool {
	for {
		answer := askQuestion(question, options, defaultVal, nil)

		if utils.ValueInSlice(strings.ToLower(answer), []string{"yes", "y"}) {
			return true
		} else if utils.ValueInSlice(strings.ToLower(answer), []string{"no", "n"}) {
			return false
		}

		fmt.Fprintf(os.Stderr, "Invalid input, try again.\n\n")
	}
}
