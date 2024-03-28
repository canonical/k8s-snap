package app

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/microcluster/state"
)

func (a *App) onStart(s *state.State) error {
	configController := controllers.NewNodeConfigurationController(a.Snap(), func(ctx context.Context) *k8s.Client {
		return k8s.RetryNewClient(ctx, a.Snap().KubernetesNodeRESTClientGetter("kube-system"))
	})
	go configController.Run(s.Context)

	return nil
}
