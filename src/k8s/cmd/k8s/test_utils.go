package k8s

import (
	"bytes"

	"github.com/canonical/k8s/pkg/k8s/client/mock"
	"github.com/spf13/cobra"
)

func setHookWithFakeClient(fakeClient *mock.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		k8sdClient = fakeClient
		return nil
	}
}

// mustSetupCLIWithFakeClient creates the CLI instance with a mocked API client.
// Returns the command, the mock and a byte buffer that contains the output of the command after run.
// See `k8s_enable_test.go` for an example.
func mustSetupCLIWithFakeClient() (*cobra.Command, *mock.Client, *bytes.Buffer) {
	cmd := NewRootCmd()
	fakeClient := &mock.Client{}
	for _, cmd := range cmd.Commands() {
		cmd.PreRunE = chainPreRunHooks(setHookWithFakeClient(fakeClient))
	}
	out := bytes.NewBufferString("")
	cmd.SetOut(out)
	return cmd, fakeClient, out
}
