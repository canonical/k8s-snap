package app

import (
	"context"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
	"github.com/canonical/microcluster/state"
)

func onStart(s *state.State) error {
	snap := snap.SnapFromContext(s.Context)

	configController := controllers.NewNodeConfigurationController(snap, func(ctx context.Context) *k8s.Client {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(3 * time.Second):
			default:
			}

			client, err := k8s.NewClient(snap.KubernetesNodeRESTClientGetter("kube-system"))
			if err != nil {
				continue
			}
			return client
		}
	})
	go configController.Run(s.Context)

	return nil
}
