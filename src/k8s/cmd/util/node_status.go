package cmdutil

import (
	"context"
	"errors"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/client/k8sd"

	"github.com/canonical/lxd/shared/api"
)

// GetNodeStatus retrieves the NodeStatus from k8sd client. If the daemon is not initialized, it exits with an error
// describing that the cluster should be bootstrapped. In case of any other errors it exits and shows the error.
func GetNodeStatus(ctx context.Context, client k8sd.Client, env ExecutionEnvironment) (status apiv1.NodeStatus, isBootstrapped bool, err error) {
	status, err = client.NodeStatus(ctx)
	if err == nil {
		return status, true, nil
	}

	if errors.As(err, &api.StatusError{}) {
		// the returned `ok` can be ignored since we're using errors.As()
		// on the same type immediately before it
		statusErr, _ := err.(api.StatusError)

		// if we get an `http.StatusServiceUnavailable` it will be (most likely) because
		// the handler we're trying to reach is not `AllowedBeforeInit` and hence we can understand that
		// the daemon is not yet initialized (this statement should be available
		// in the `statusErr.Error()` explicitly but for the sake of decoupling we don't rely on that)
		if statusErr.Status() == http.StatusServiceUnavailable {
			return status, false, err
		}
	}

	return status, true, err
}
