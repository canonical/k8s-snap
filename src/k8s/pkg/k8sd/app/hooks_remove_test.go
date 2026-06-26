// package app_test is used (and not app) to avoid an import cycle: testenv imports
// package app, so internal test files that also import testenv would create a cycle.
// export_test.go bridges the gap by re-exporting the unexported symbols needed here.

package app_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/app"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	testenv "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/microcluster/v2/cluster"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
)

// TestOnPreRemoveNodeAbsentFromDB tests that, when the local node has no entry
// in the microcluster DB, the PENDING wait loop exits on the first iteration
// (NotFound -> notPending = true) rather than spinning until context timeout.
// The function must complete well within a 5-second deadline.
func TestOnPreRemoveNodeAbsentFromDB(t *testing.T) {
	testenv.WithState(t, func(ctx context.Context, s state.State) {
		g := NewWithT(t)

		// Remove the local node from the cluster DB so GetCoreClusterMember
		// returns NotFound, which is the scenario this fix targets.
		err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			member, err := cluster.GetCoreClusterMember(ctx, tx, s.Name())
			if err != nil {
				return err
			}
			return cluster.DeleteCoreClusterMember(ctx, tx, member.Address)
		})
		g.Expect(err).ToNot(HaveOccurred())

		snap := &snapmock.Snap{Mock: snapmock.Mock{}}
		a := app.NewTestApp(snap)

		// Tight deadline: without the fix the loop spins every second until context
		// expires. With the fix the loop exits on the first DB check.
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err = app.OnPreRemove(a, ctx, s, true)
		g.Expect(err).ToNot(HaveOccurred())

		// Context must still be valid. A spin-loop would have exhausted it.
		g.Expect(ctx.Err()).To(BeNil(), "context expired: PENDING wait loop likely timed out")
	})
}
