package app

import (
	"context"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/microcluster/state"
)

func (a *App) onStart(s *state.State) error {
	configController := controllers.NewNodeConfigurationController(a.Snap(), func(ctx context.Context) *k8s.Client {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(3 * time.Second):
			}

			client, err := k8s.NewClient(a.Snap().KubernetesNodeRESTClientGetter("kube-system"))
			if err != nil {
				continue
			}
			return client
		}
	})
	go configController.Run(s.Context)

	return nil
}
