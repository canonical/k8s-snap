package k8sd_test

import (
	"context"
	"testing"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/client/k8sd"
	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
)

func TestNodeStatus(t *testing.T) {
	snap := &mock.Snap{}

	WithMicrocluster(t, app.Config{
		StateDir: t.TempDir(),
		Snap:     snap,
	}, &state.Hooks{
		// nullify onStart and postBootstrap hooks
		OnStart:       func(ctx context.Context, s state.State) error { return nil },
		PostBootstrap: func(ctx context.Context, s state.State, initConfig map[string]string) error { return nil },
	}, func(ctx context.Context, address string, app *app.App, client k8sd.Client) {

		t.Run("NotInitialized", func(t *testing.T) {
			g := NewWithT(t)
			_, initilized, err := client.NodeStatus(context.Background())
			g.Expect(err).To(BeNil())
			g.Expect(initilized).To(BeFalse())
		})

		t.Run("Initialized", func(t *testing.T) {
			g := NewWithT(t)
			_, err := client.BootstrapCluster(ctx, apiv1.BootstrapClusterRequest{
				Name:    "t1",
				Address: address,
				Timeout: 30 * time.Second,
			})
			g.Expect(err).To(BeNil())

			resp, initialized, err := client.NodeStatus(context.Background())
			g.Expect(err).To(BeNil())
			g.Expect(initialized).To(BeTrue())
			g.Expect(resp.NodeStatus.Name).To(Equal("t1"))
		})
	})
}
