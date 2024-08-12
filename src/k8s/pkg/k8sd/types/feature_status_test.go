package types_test

import (
	"testing"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
)

func TestK8sdFeatureStatusToAPI(t *testing.T) {
	k8sdFS := types.FeatureStatus{
		Enabled:   true,
		Message:   "message",
		Version:   "version",
		UpdatedAt: time.Now(),
	}

	apiFS := k8sdFS.ToAPI()
	g := NewWithT(t)
	g.Expect(apiFS.Enabled).To(Equal(k8sdFS.Enabled))
	g.Expect(apiFS.Message).To(Equal(k8sdFS.Message))
	g.Expect(apiFS.Version).To(Equal(k8sdFS.Version))
	g.Expect(apiFS.UpdatedAt).To(Equal(k8sdFS.UpdatedAt))
}

func TestAPIFeatureStatusToK8sd(t *testing.T) {
	apiFS := apiv1.FeatureStatus{
		Enabled:   true,
		Message:   "message",
		Version:   "version",
		UpdatedAt: time.Now(),
	}

	k8sdFS := types.FeatureStatusFromAPI(apiFS)
	g := NewWithT(t)
	g.Expect(k8sdFS.Enabled).To(Equal(apiFS.Enabled))
	g.Expect(k8sdFS.Message).To(Equal(apiFS.Message))
	g.Expect(k8sdFS.Version).To(Equal(apiFS.Version))
	g.Expect(k8sdFS.UpdatedAt).To(Equal(apiFS.UpdatedAt))
}
