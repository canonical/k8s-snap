package k8sd_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/client/k8sd"
	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/canonical/microcluster/v2/state"
)

var (
	// nextIdx is used to pick different listen ports for each microcluster instance
	nextIdx int
)

// WithMicrocluster can be used to run isolated tests against a microcluster instance.
// WithMicrocluster accepts app configuration and can optionally override any microcluster hooks.
// WithMicrocluster accepts the test code as a function, and passes a context, a local node address that can be used for microcluster, the app object and a k8sd.Client.
//
// Example usage:
//
//	func TestNodeStatus(t *testing.T) {
//		snap := &mock.Snap{}
//		WithMicrocluster(t, app.Config{
//			StateDir: t.TempDir(),
//			Snap:     snap,
//		}, &state.Hooks{
//			// nullify onStart and postBootstrap hooks
//			OnStart:       func(ctx context.Context, s state.State) error { return nil },
//			PostBootstrap: func(ctx context.Context, s state.State, initConfig map[string]string) error { return nil },
//		}, func(ctx context.Context, address string, app *app.App, client k8sd.Client) {
//			g := NewWithT(t)
//
//			_, initilized, err := client.NodeStatus(context.Background())
//			g.Expect(err).To(BeNil())
//			g.Expect(initilized).To(BeFalse())
//
//			_, err = client.BootstrapCluster(ctx, apiv1.BootstrapClusterRequest{
//				Name:    "t1",
//				Address: address,
//				Timeout: 30 * time.Second,
//			})
//			g.Expect(err).To(BeNil())
//
//			resp, initialized, err := client.NodeStatus(context.Background())
//			g.Expect(err).To(BeNil())
//			g.Expect(initialized).To(BeTrue())
//			g.Expect(resp.NodeStatus.Name).To(Equal("t1"))
//		})
//	}
func WithMicrocluster(t *testing.T, config app.Config, hooks *state.Hooks, f func(context.Context, string, *app.App, k8sd.Client)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := app.New(config)
	if err != nil {
		t.Fatalf("failed to create microcluster app: %v", err)
	}

	client, err := k8sd.New(config.StateDir, "")
	if err != nil {
		t.Fatalf("failed to create k8sd client: %v", err)
	}

	// app.Run() is blocking, start in a goroutine.
	go func() {
		if err := app.Run(ctx, hooks); err != nil {
			t.Logf("microcluster app failed: %v", err)
		}
	}()

	if err := app.MicroCluster().Ready(ctx); err != nil {
		t.Fatalf("microcluster app was not ready in time: %v", err)
	}

	nextIdx++
	f(ctx, fmt.Sprintf("127.0.0.1:%d", 52030+nextIdx), app, client)

	// cancel context to stop the microcluster instance, and wait for it to shutdown
	cancel()
}
