package types

import (
	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

// RefreshStatus represents the status of a snap refresh operation.
// This is a partial struct derived from the Change struct used by the snapd API.
type RefreshStatus struct {
	// Status is the current status of the operation.
	Status string `json:"status"`
	// Ready indicates whether the operation has completed.
	Ready bool `json:"ready"`
	// Err contains an error message if the operation failed.
	Err string `json:"err,omitempty"`
}

func (r RefreshStatus) ToAPI() apiv1.SnapRefreshStatusResponse {
	return apiv1.SnapRefreshStatusResponse{
		Status:       r.Status,
		Completed:    r.Ready,
		ErrorMessage: r.Err,
	}
}
