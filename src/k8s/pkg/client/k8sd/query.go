package k8sd

import (
	"context"
	"fmt"
	"strings"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/lxd/shared/api"
)

// query is a helper method to wrap common error checking and response handling.
func query[T any](ctx context.Context, c *k8sd, method string, path string, in any, out *T) (T, error) {
	if err := c.client.Query(ctx, method, apiv1.K8sdAPIVersion, api.NewURL().Path(strings.Split(path, "/")...), in, out); err != nil {
		var zero T
		return zero, fmt.Errorf("failed to %s /%s: %w", method, path, err)
	}
	return *out, nil
}
