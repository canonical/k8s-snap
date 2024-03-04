package k8s

import (
	"fmt"

	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

var (
	revokeAuthTokenCmdOpts struct {
		token string
	}
)

func newRevokeAuthTokenCmd() *cobra.Command {
	revokeAuthTokenCmd := &cobra.Command{
		Use:               "revoke-auth-token --token <token>",
		Short:             "Revoke an auth token for Kubernetes",
		Hidden:            true,
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			if err := k8sdClient.RevokeAuthToken(cmd.Context(), revokeAuthTokenCmdOpts.token); err != nil {
				return fmt.Errorf("Could not revoke auth token: %w", err)
			}

			return nil
		},
	}
	revokeAuthTokenCmd.Flags().StringVar(&revokeAuthTokenCmdOpts.token, "token", "", "Token")
	return revokeAuthTokenCmd
}
