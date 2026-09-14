// package api_test is used (and not api) to avoid an import cycle: testenv imports
// pkg/k8sd/app, which imports pkg/k8sd/api, so internal test files that also
// import testenv would create a cycle. export_test.go bridges the gap by
// re-exporting the unexported symbols needed here.

package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/api"
	testenv "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
)

// TestRemoveNodeFromMicroclusterAbsentFromDB tests that, when the target node has no entry
// in the microcluster DB, the PENDING wait loop exits on the first iteration
// (NotFound -> notPending = true) rather than spinning until context timeout.
// The function must complete well within a 5-second deadline.
func TestRemoveNodeFromMicroclusterAbsentFromDB(t *testing.T) {
	testenv.WithState(t, func(ctx context.Context, s state.State) {
		g := NewWithT(t)

		// Tight deadline: without the fix the loop spins every second until context
		// expires. With the fix the loop exits on the first DB check.
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err := api.RemoveNodeFromMicrocluster(ctx, s, "never-joined-node", false)
		g.Expect(err).ToNot(HaveOccurred())

		// Context must still be valid. A spin-loop would have exhausted it.
		g.Expect(ctx.Err()).To(BeNil(), "context expired: PENDING wait loop likely spun to timeout")
	})
}
