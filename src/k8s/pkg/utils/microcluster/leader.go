package microcluster

import (
	"fmt"

	"github.com/canonical/microcluster/v2/state"
)

func IsLeader(s state.State) (bool, error) {
	leaderClient, err := s.Leader()
	if err != nil {
		return false, fmt.Errorf("failed to get leader client: %w", err)
	}

	leaderURL := leaderClient.URL()
	nodeURL := s.Address()

	if leaderURL.Hostname() != nodeURL.Hostname() {
		return false, nil
	}

	return true, nil
}
