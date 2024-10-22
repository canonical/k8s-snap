package types

import (
	"fmt"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

// RefreshOpts controls the target version of the snap during a refresh.
type RefreshOpts struct {
	// LocalPath refreshes the snap using a local snap archive, e.g. "/path/to/k8s.snap".
	LocalPath string `json:"localPath"`
	// Channel refreshes the snap to track a specific channel, e.g. "latest/edge".
	Channel string `json:"channel"`
	// Revision refreshes the snap to a specific revision, e.g. "722".
	Revision string `json:"revision"`
}

func RefreshOptsFromAPI(req apiv1.SnapRefreshRequest) (RefreshOpts, error) {
	optsMap := map[string]string{
		"localPath": req.LocalPath,
		"channel":   req.Channel,
		"revision":  req.Revision,
	}

	// Make sure only one of the options is set.
	alreadySet := false
	for _, v := range optsMap {
		if alreadySet && v != "" {
			return RefreshOpts{}, fmt.Errorf("only one of localPath, channel or revision can be specified")
		}
		if v != "" {
			alreadySet = true
		}
	}

	switch {
	case req.LocalPath != "":
		return RefreshOpts{LocalPath: req.LocalPath}, nil
	case req.Channel != "":
		return RefreshOpts{Channel: req.Channel}, nil
	case req.Revision != "":
		return RefreshOpts{Revision: req.Revision}, nil
	}
	return RefreshOpts{}, fmt.Errorf("empty snap refresh target")
}
