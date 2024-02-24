package k8s

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
	v1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	bootstrapCmdOpts struct {
		interactive bool
		timeout     time.Duration
	}

	bootstrapCmdErrorMsgs = map[error]string{
		apiv1.ErrUnknown:             "An unknown error occured while bootstrapping the cluster:\n",
		apiv1.ErrAlreadyBootstrapped: "K8s cluster already bootstrapped.",
	}
)

func newBootstrapCmd() *cobra.Command {
	bootstrapCmd := &cobra.Command{
		Use:               "bootstrap",
		Short:             "Bootstrap a k8s cluster on this node.",
		Long:              "Initialize the necessary folders, permissions, service arguments, certificates and start up the Kubernetes services.",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, bootstrapCmdErrorMsgs)

			if k8sdClient.IsBootstrapped(cmd.Context()) {
				return v1.ErrAlreadyBootstrapped
			}

			const minTimeout = 3 * time.Second
			if bootstrapCmdOpts.timeout < minTimeout {
				cmd.PrintErrf("Timeout %v is less than minimum of %v. Using the minimum %v instead.\n", bootstrapCmdOpts.timeout, minTimeout, minTimeout)
				bootstrapCmdOpts.timeout = minTimeout
			}

			config := apiv1.BootstrapConfig{}
			if bootstrapCmdOpts.interactive {
				config = getConfigInteractively(cmd.Context())
			} else {
				config.SetDefaults()
			}

			fmt.Println("Bootstrapping the cluster. This may take some seconds...")
			cluster, err := k8sdClient.Bootstrap(cmd.Context(), config)
			if err != nil {
				return fmt.Errorf("failed to bootstrap cluster: %w", err)
			}

			fmt.Printf("Bootstrapped k8s cluster on %q (%s).\n", cluster.Name, cluster.Address)
			return nil
		},
	}

	bootstrapCmd.PersistentFlags().BoolVar(&bootstrapCmdOpts.interactive, "interactive", false,
		"Interactively configure the most important cluster options.")
	bootstrapCmd.PersistentFlags().DurationVar(&bootstrapCmdOpts.timeout, "timeout", 90*time.Second, "The max time to wait for k8s to bootstrap.")

	return bootstrapCmd
}

func getConfigInteractively(ctx context.Context) apiv1.BootstrapConfig {
	config := apiv1.BootstrapConfig{}
	config.SetDefaults()

	components := askQuestion("Which components would you like to enable?", []string{}, strings.Join(config.Components, ", "))
	// TODO: Validate components
	config.Components = strings.Split(strings.ReplaceAll(components, " ", ""), ",")

	config.ClusterCIDR = askQuestion("Please set the Cluster CIDR?", nil, config.ClusterCIDR)

	rbac := askBool("Enable Role Based Access Control (RBAC)?", []string{"yes", "no"}, "yes")
	*config.EnableRBAC = rbac
	return config
}

func askQuestion(question string, options []string, defaultVal string) string {
	if options != nil {
		question = fmt.Sprintf("%s (%s)", question, strings.Join(options, ", "))
	}
	if defaultVal != "" {
		question = fmt.Sprintf("%s [%s]:", question, defaultVal)
	}
	question = fmt.Sprintf("%s ", question)

	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stdout, question)
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	s = strings.TrimSpace(s)

	if s == "" {
		return defaultVal
	}
	return s
}

// askBool asks a question and expect a yes/no answer.
func askBool(question string, options []string, defaultVal string) bool {
	for {
		answer := askQuestion(question, options, defaultVal)

		if utils.ValueInSlice(strings.ToLower(answer), []string{"yes", "y"}) {
			return true
		} else if utils.ValueInSlice(strings.ToLower(answer), []string{"no", "n"}) {
			return false
		}

		fmt.Fprintf(os.Stderr, "Invalid input, try again.\n\n")
	}
}
