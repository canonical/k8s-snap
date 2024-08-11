package types

import (
	"time"

	apiv1 "github.com/canonical/k8s-snap-api-v1/api/v1"
)

// FeatureStatus encapsulates the deployment status of a feature.
type FeatureStatus struct {
	// Enabled shows whether or not the deployment of manifests for a status was successful.
	Enabled bool
	// Message contains information about the status of a feature. It is only supposed to be human readable and informative and should not be programmatically parsed.
	Message string
	// Version shows the version of the deployed feature.
	Version string
	// UpdatedAt shows when the last update was done.
	UpdatedAt time.Time
}

func (f FeatureStatus) ToAPI() apiv1.FeatureStatus {
	return apiv1.FeatureStatus{
		Enabled:   f.Enabled,
		Message:   f.Message,
		Version:   f.Version,
		UpdatedAt: f.UpdatedAt,
	}
}

func FeatureStatusFromAPI(apiFS apiv1.FeatureStatus) FeatureStatus {
	return FeatureStatus{
		Enabled:   apiFS.Enabled,
		Message:   apiFS.Message,
		Version:   apiFS.Version,
		UpdatedAt: apiFS.UpdatedAt,
	}
}
