package app

import (
	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/state"
)

func onStart(s *state.State) error {
	snap := snap.SnapFromContext(s.Context)

	configController := controllers.NewNodeConfigurationController(snap)
	go configController.Run(s.Context)

	return nil
}
