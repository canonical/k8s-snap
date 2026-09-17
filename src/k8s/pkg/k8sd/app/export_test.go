// Tests that need testenv.WithState must live in package app_test to avoid an
// import cycle (testenv imports package app), so this file bridges the gap by
// re-exporting the symbols they need.

package app

import (
	"context"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

func NewTestApp(s snap.Snap) *App {
	return &App{snap: s}
}

func OnPreRemove(a *App, ctx context.Context, s state.State, force bool) error {
	return a.onPreRemove(ctx, s, force)
}
