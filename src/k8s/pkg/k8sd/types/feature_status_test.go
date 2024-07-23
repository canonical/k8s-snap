package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

func TestK8sdFeatureStatusToAPI(t *testing.T) {
	k8sdFS := types.FeatureStatus{
		Enabled:   true,
		Message:   "message",
		Version:   "version",
		UpdatedAt: time.Now(),
	}

	apiFS, err := k8sdFS.ToAPI()
	require.NoError(t, err)
	assert.Equal(t, k8sdFS.Enabled, apiFS.Enabled)
	assert.Equal(t, k8sdFS.Message, apiFS.Message)
	assert.Equal(t, k8sdFS.Version, apiFS.Version)
	assert.Equal(t, k8sdFS.UpdatedAt, apiFS.UpdatedAt)
}

func TestAPIFeatureStatusToK8sd(t *testing.T) {
	apiFS := apiv1.FeatureStatus{
		Enabled:   true,
		Message:   "message",
		Version:   "version",
		UpdatedAt: time.Now(),
	}

	k8sdFS, err := types.FeatureStatusFromAPI(apiFS)
	require.NoError(t, err)
	assert.Equal(t, apiFS.Enabled, k8sdFS.Enabled)
	assert.Equal(t, apiFS.Message, k8sdFS.Message)
	assert.Equal(t, apiFS.Version, k8sdFS.Version)
	assert.Equal(t, apiFS.UpdatedAt, k8sdFS.UpdatedAt)
}
