package app

import (
	"github.com/canonical/microcluster/state"
)

func (a *App) onStart(s *state.State) error {
	// start a goroutine to mark the node as running
	go a.markNodeReady(s.Context, s)

	// start node config controller
	if a.nodeConfigController != nil {
		go a.nodeConfigController.Run(s.Context)
	}

	return nil
}
