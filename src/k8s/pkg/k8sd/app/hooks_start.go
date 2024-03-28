package app

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/microcluster/state"
)

func onStart(s *state.State) error {
	snap := snap.SnapFromContext(s.Context)

	configController := controllers.NewNodeConfigurationController(snap, func(ctx context.Context) *k8s.Client {
		return k8s.RetryNewClient(ctx, snap.KubernetesNodeRESTClientGetter("kube-system"))
	})
	go configController.Run(s.Context)

	return nil
}
