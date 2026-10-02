// Tests that need testenv.WithState must live in package api_test to avoid an
// import cycle (testenv -> app -> api), so this file bridges the gap by
// re-exporting the symbols they need.

package api

import (
	"context"

	"github.com/canonical/microcluster/v2/state"
)

var RemoveNodeFromMicrocluster = func(ctx context.Context, s state.State, nodeName string, force bool) error {
	return removeNodeFromMicrocluster(ctx, s, nodeName, force)
}
