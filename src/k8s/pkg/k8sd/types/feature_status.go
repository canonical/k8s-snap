package types

import (
	"time"

	apiv1 "github.com/canonical/k8s/api/v1"
)

type FeatureStatus struct {
	Enabled   bool
	Message   string
	Version   string
	Timestamp time.Time
}

func (f FeatureStatus) ToAPI() (apiv1.FeatureStatus, error) {
	return apiv1.FeatureStatus{
		Enabled:   f.Enabled,
		Message:   f.Message,
		Version:   f.Version,
		Timestamp: f.Timestamp,
	}, nil
}

func FeatureStatusFromAPI(apiFS apiv1.FeatureStatus) (FeatureStatus, error) {
	return FeatureStatus{
		Enabled:   apiFS.Enabled,
		Message:   apiFS.Message,
		Version:   apiFS.Version,
		Timestamp: apiFS.Timestamp,
	}, nil
}
