package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/canonical/microcluster/v3/state"
)

const (
	// microclusterDatabaseInitTimeout is the timeout for microcluster database initialization operations.
	microclusterDatabaseInitTimeout = 3 * time.Second
	// microclusterDatabaseShutdownTimeout is the timeout for microcluster database shutdown operations.
	microclusterDatabaseShutdownTimeout = 3 * time.Second
)

var (
	// nextIdx is used to pick different listen ports for each microcluster instance.
	nextIdx int
)

// DB is an interface for the internal microcluster DB type.
type DB interface {
	Transaction(ctx context.Context, f func(context.Context, *sql.Tx) error) error
}

// WithDB can be used to run isolated tests against the microcluster database.
//
// Example usage:
//
//	func TestKubernetesAuthTokens(t *testing.T) {
//		t.Run("ValidToken", func(t *testing.T) {
//			g := NewWithT(t)
//			WithDB(t, func(ctx context.Context, db DB) {
//				err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
//					token, err := database.GetOrCreateToken(ctx, tx, "user1", []string{"group1", "group2"})
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
func WithDB(t *testing.T, f func(context.Context, DB)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := app.New(app.Config{
		StateDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("failed to create microcluster app: %v", err)
	}

	databaseCh := make(chan DB, 1)
	doneCh := make(chan error, 1)
	defer close(databaseCh)
	defer close(doneCh)

	// app.Run() is blocking, so we get the database handle through a channel
	go func() {
		doneCh <- app.Run(ctx, &state.Hooks{
			PostBootstrap: func(ctx context.Context, s state.State, initConfig map[string]string) error {
				databaseCh <- s.Database()
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
	case db := <-databaseCh:
		f(ctx, db)
	}

	// cancel context to stop the microcluster instance, and wait for it to shutdown
	cancel()
	select {
	case <-doneCh:
	case <-time.After(microclusterDatabaseShutdownTimeout):
		t.Fatalf("timed out waiting for microcluster to shutdown")
	}
}
