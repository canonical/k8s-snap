package k8sd

import (
	"context"
	"fmt"
	"strings"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcluster/v2/rest/types"
)

// query is a helper method for sending requests to the k8sd client with common error checking and automatic retries.
// It retries on temporary microcluster errors and returns the deserialized response.
func query[T any](ctx context.Context, c *k8sd, method, path string, in any, out *T) (T, error) {
	var result T
	if out == nil {
		return result, fmt.Errorf("out must be a non-nil pointer")
	}

	retryErr := control.WaitUntilReady(ctx, func() (bool, error) {
		err := c.client.Query(ctx, method, apiv1.K8sdAPIVersion, api.NewURL().Path(strings.Split(path, "/")...), in, out)
		if err != nil {
			if isTemporary(err) {
				log.FromContext(ctx).Info("Temporary error from k8sd: %v", err)
				return false, nil
			}
			if bootstrapPending(err) {
				return false, fmt.Errorf(
					"the cluster hasn't been initialized yet. "+
						" Attempted k8sd query: %s /%s: %w", method, path, err)
			}
			return false, fmt.Errorf("failed to %s /%s: %w", method, path, err)
		}
		return true, nil
	})

	if retryErr != nil {
		return result, fmt.Errorf("failed after potential retry: %w", retryErr)
	}

	return *out, nil
}

// isTemporary checks if an error is temporary and should be retried.
// This function is tightly coupled with the error messages returned by microcluster and
// should not contain any generic error checks.
func isTemporary(err error) bool {
	if strings.Contains(err.Error(), string(types.DatabaseStarting)) ||
		strings.Contains(err.Error(), "Database is not yet ready") {
		return true
	}
	return false
}

// bootstrapPending checks if an error was caused by the fact that the cluster
// has not been initialized yet.
func bootstrapPending(err error) bool {
	return strings.Contains(err.Error(), "Database is not yet initialized")
}
