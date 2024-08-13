package k8sd_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/client/k8sd"
	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
)

func TestNodeJoin(t *testing.T) {
	snap1 := &mock.Snap{}
	snap2 := &mock.Snap{}

	// nullify onStart and postBootstrap hooks
	hooks := &state.Hooks{
		OnStart:       func(context.Context, state.State) error { return nil },
		PostBootstrap: func(context.Context, state.State, map[string]string) error { return nil },
	}

	t.Run("JoinOne", func(t *testing.T) {
		g := NewWithT(t)
		WithMicrocluster(t, app.Config{StateDir: t.TempDir(), Snap: snap1}, hooks, func(ctx1 context.Context, address1 string, mc1 *app.App, client1 k8sd.Client) {
			_, err := client1.BootstrapCluster(ctx1, apiv1.BootstrapClusterRequest{
				Name:    "t1",
				Address: address1,
				Timeout: 30 * time.Second,
			})
			g.Expect(err).To(BeNil())

			resp, err := client1.GetJoinToken(ctx1, apiv1.GetJoinTokenRequest{
				Name:   "t2",
				Worker: false,
			})
			g.Expect(err).To(BeNil())

			WithMicrocluster(t, app.Config{StateDir: t.TempDir(), Snap: snap2}, hooks, func(ctx2 context.Context, address2 string, mc2 *app.App, client2 k8sd.Client) {
				g.Expect(err).To(BeNil())

				fmt.Println(address1, address2)
				err = client2.JoinCluster(ctx2, apiv1.JoinClusterRequest{
					Name:    "t2",
					Address: address2,
					Token:   resp.EncodedToken,
					Config:  "{}",
				})
				g.Expect(err).To(BeNil())
			})
		})
	})
}
