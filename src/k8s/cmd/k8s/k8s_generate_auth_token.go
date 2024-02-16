package k8s

import (
	"fmt"

	"github.com/canonical/k8s/cmd/k8s/errors"

	"github.com/spf13/cobra"
)

var (
	generateAuthTokenCmdOpts struct {
		username string
		groups   []string
	}
)

func newGenerateAuthTokenCmd() *cobra.Command {
	generateAuthTokenCmd := &cobra.Command{
		Use:               "generate-auth-token --username <user> [--groups <group1>,<group2>]",
		Short:             "Generate an auth token for Kubernetes",
		Hidden:            true,
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			token, err := k8sdClient.GenerateAuthToken(cmd.Context(), generateAuthTokenCmdOpts.username, generateAuthTokenCmdOpts.groups)
			if err != nil {
				return fmt.Errorf("could not generate auth token: %w", err)
			}
			fmt.Println(token)

			return nil
		},
	}
	generateAuthTokenCmd.Flags().StringVar(&generateAuthTokenCmdOpts.username, "username", "", "Username")
	generateAuthTokenCmd.Flags().StringSliceVar(&generateAuthTokenCmdOpts.groups, "groups", nil, "Groups")
	return generateAuthTokenCmd
}
