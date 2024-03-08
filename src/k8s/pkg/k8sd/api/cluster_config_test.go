package api

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func DefaultConfig() types.ClusterConfig {
	config := types.ClusterConfig{}
	config.SetDefaults()
	return config
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		oldConfig     types.ClusterConfig
		newConfig     types.ClusterConfig
		expectedError string
	}{
		{
			name: "Disable network should not work before load-balancer is disabled",
			oldConfig: types.ClusterConfig{
				Network: types.Network{
					Enabled: vals.Pointer(true),
				},
				LoadBalancer: types.LoadBalancer{
					Enabled: vals.Pointer(true),
				},
			},
			newConfig: types.ClusterConfig{
				Network: types.Network{
					Enabled: vals.Pointer(false),
				},
				LoadBalancer: types.LoadBalancer{
					Enabled: vals.Pointer(true),
				},
			},
			expectedError: "load-balancer must be disabled",
		},
		{
			name: "Disable network should work if load-balancer is also disabled in same request",
			oldConfig: types.ClusterConfig{
				Network: types.Network{
					Enabled: vals.Pointer(true),
				},
				LoadBalancer: types.LoadBalancer{
					Enabled: vals.Pointer(true),
				},
			},
			newConfig: types.ClusterConfig{
				Network: types.Network{
					Enabled: vals.Pointer(false),
				},
				LoadBalancer: types.LoadBalancer{
					Enabled: vals.Pointer(false),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			newConfig, err := types.MergeClusterConfig(DefaultConfig(), tt.newConfig)
			g.Expect(err).To(BeNil())

			err = validateConfig(tt.oldConfig, newConfig)
			if tt.expectedError == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring(tt.expectedError))
			}
		})
	}
}
