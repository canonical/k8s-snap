package microcluster_testenv

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/canonical/microcluster/v2/state"
)

const (
	// microclusterDatabaseInitTimeout is the timeout for microcluster database initialization operations.
	microclusterDatabaseInitTimeout = 3 * time.Second
	// microclusterDatabaseShutdownTimeout is the timeout for microcluster database shutdown operations.
	microclusterDatabaseShutdownTimeout = 3 * time.Second
)

// nextIdx is used to pick different listen ports for each microcluster instance.
var nextIdx int

// WithState can be used to run isolated tests against the microcluster database.
// The Database() can be accessed by calling s.Database().
//
// Example usage:
//
//	func TestKubernetesAuthTokens(t *testing.T) {
//		t.Run("ValidToken", func(t *testing.T) {
//			g := NewWithT(t)
//			WithState(t, func(ctx context.Context, s state.State) {
//				err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
//					token, err := s.Database().GetOrCreateToken(ctx, tx, "user1", []string{"group1", "group2"})
//					if !g.Expect(err).To(Not(HaveOccurred())) {
//						return err
//					}
//					g.Expect(token).To(Not(BeEmpty()))
//					return nil
//				})
//				g.Expect(err).To(Not(HaveOccurred()))
//			})
//		})
//	}
func WithState(t *testing.T, f func(context.Context, state.State)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := app.New(app.Config{
		StateDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("failed to create microcluster app: %v", err)
	}

	stateChan := make(chan state.State, 1)
	doneCh := make(chan error, 1)
	defer close(stateChan)
	defer close(doneCh)

	// app.Run() is blocking, so we get the state handle through a channel
	go func() {
		doneCh <- app.Run(ctx, &state.Hooks{
			PostBootstrap: func(ctx context.Context, s state.State, initConfig map[string]string) error {
				stateChan <- s
				return nil
			},
			OnStart: func(ctx context.Context, s state.State) error {
				return nil
			},
		})
	}()

	if err := app.MicroCluster().Ready(ctx); err != nil {
		t.Fatalf("microcluster app was not ready in time: %v", err)
	}

	nextIdx++
	if err := app.MicroCluster().NewCluster(ctx, fmt.Sprintf("test-%d", nextIdx), fmt.Sprintf("127.0.0.1:%d", 51030+nextIdx), nil); err != nil {
		t.Fatalf("microcluster app failed to bootstrap: %v", err)
	}

	select {
	case err := <-doneCh:
		if err != nil {
			t.Fatalf("microcluster app failed: %v", err)
		}
	default:
	}

	select {
	case <-time.After(microclusterDatabaseInitTimeout):
		t.Fatalf("timed out waiting for microcluster to start")
	case state := <-stateChan:
		f(ctx, state)
	}

	// cancel context to stop the microcluster instance, and wait for it to shutdown
	cancel()
	select {
	case <-doneCh:
	case <-time.After(microclusterDatabaseShutdownTimeout):
		t.Fatalf("timed out waiting for microcluster to shutdown")
	}
}
