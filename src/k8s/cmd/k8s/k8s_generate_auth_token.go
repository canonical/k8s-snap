package k8s

import (
	"fmt"

	"github.com/canonical/k8s/pkg/k8s/client"

	"github.com/spf13/cobra"
)

var (
	generateAuthTokenCmdOpts struct {
		username string
		groups   []string
	}

	generateAuthTokenCmd = &cobra.Command{
		Use:    "generate-auth-token --username <user> [--groups <group1>,<group2>]",
		Short:  "Generate an auth token for Kubernetes",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := client.NewClient(cmd.Context(), client.ClusterOpts{
				StateDir: clusterCmdOpts.stateDir,
				Verbose:  rootCmdOpts.logVerbose,
				Debug:    rootCmdOpts.logDebug,
			})
			if err != nil {
				return fmt.Errorf("failed to create cluster client: %w", err)
			}

			token, err := client.GenerateAuthToken(cmd.Context(), generateAuthTokenCmdOpts.username, generateAuthTokenCmdOpts.groups)
			if err != nil {
				return fmt.Errorf("could not generate auth token: %w", err)
			}
			fmt.Println(token)

			return nil
		},
	}
)

func init() {
	generateAuthTokenCmd.Flags().StringVar(&generateAuthTokenCmdOpts.username, "username", "", "Username")
	generateAuthTokenCmd.Flags().StringSliceVar(&generateAuthTokenCmdOpts.groups, "groups", nil, "Groups")

	rootCmd.AddCommand(generateAuthTokenCmd)
}
