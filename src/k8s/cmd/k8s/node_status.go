package k8s

import (
	"errors"
	"net/http"

	"github.com/spf13/cobra"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/client/k8sd"

	"github.com/canonical/lxd/shared/api"
)

func GetNodeStatus(client k8sd.Client, cmd *cobra.Command, env cmdutil.ExecutionEnvironment) apiv1.NodeStatus {
	status, err := client.NodeStatus(cmd.Context())
	if err == nil {
		return status
	}

	if errors.As(err, &api.StatusError{}) {
		// the returned `ok` can be ignored since we're using errors.As()
		// on the same type immediately before it
		statusErr, _ := err.(api.StatusError)
		if statusErr.Status() == http.StatusServiceUnavailable {
			cmd.PrintErrln("Error: The node is not part of a Kubernetes cluster. You can bootstrap a new cluster with:\n\n  sudo k8s bootstrap")
			env.Exit(1)
			return status
		}
	}

	cmd.PrintErrf("Error: Failed to retrieve the node status.\n\nThe error was: %v\n", err)
	env.Exit(1)
	return status
}
