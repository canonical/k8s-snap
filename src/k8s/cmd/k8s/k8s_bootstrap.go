package k8s

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8s/client"
	"github.com/spf13/cobra"
)

var (
	bootstrapCmdOpts struct {
		interactive bool
	}

	boostrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap a k8s cluster on this node.",
		Long:  "Initialize the necessary folders, permissions, service arguments, certificates and start up the Kubernetes services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				StateDir: clusterCmdOpts.stateDir,
				Verbose:  rootCmdOpts.logVerbose,
				Debug:    rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			if c.IsBootstrapped(cmd.Context()) {
				return fmt.Errorf("k8s cluster already bootstrapped")
			}

			config := apiv1.BootstrapConfig{}
			if bootstrapCmdOpts.interactive {
				config = getConfigInteractively(cmd.Context())
			} else {
				config.SetDefaults()
			}

			if !config.IsValidCIDR() {
				return fmt.Errorf("invalid cluster CIDR: %s", config.ClusterCIDR)
			}

			cluster, err := c.Bootstrap(cmd.Context(), config)
			if err != nil {
				return fmt.Errorf("failed to initialize k8s cluster: %w", err)
			}

			fmt.Printf("Bootstrapped k8s cluster on %q (%s).\n", cluster.Name, cluster.Address)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(boostrapCmd)
	boostrapCmd.PersistentFlags().BoolVar(&bootstrapCmdOpts.interactive, "interactive", false,
		"Interactively configure the most important cluster options.")
}

func getConfigInteractively(ctx context.Context) apiv1.BootstrapConfig {
	config := apiv1.BootstrapConfig{}
	config.SetDefaults()

	components := askQuestion("Which components would you like to enable?", componentList, strings.Join(config.Components, ", "))
	// TODO: Validate components
	config.Components = strings.Split(strings.ReplaceAll(components, " ", ""), ",")

	config.ClusterCIDR = askQuestion("Please set the Cluster CIDR?", nil, config.ClusterCIDR)
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
