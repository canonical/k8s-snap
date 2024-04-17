package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestValidateCIDR(t *testing.T) {
	for _, tc := range []struct {
		cidr      string
		expectErr bool
	}{
		{cidr: "10.1.0.0/16"},
		{cidr: "2001:0db8::/32"},
		{cidr: "10.1.0.0/16,2001:0db8::/32"},
		{cidr: "", expectErr: true},
		{cidr: "bananas", expectErr: true},
		{cidr: "fd01::/64,fd02::/64,fd03::/64", expectErr: true},
	} {
		t.Run(tc.cidr, func(t *testing.T) {
			t.Run("Pod", func(t *testing.T) {
				g := NewWithT(t)
				config := types.ClusterConfig{
					Network: types.Network{
						PodCIDR:     vals.Pointer(tc.cidr),
						ServiceCIDR: vals.Pointer("10.1.0.0/16"),
					},
				}
				err := config.Validate()
				if tc.expectErr {
					g.Expect(err).To(HaveOccurred())
				} else {
					g.Expect(err).To(BeNil())
				}
			})
			t.Run("Service", func(t *testing.T) {
				g := NewWithT(t)
				config := types.ClusterConfig{
					Network: types.Network{
						PodCIDR:     vals.Pointer("10.1.0.0/16"),
						ServiceCIDR: vals.Pointer(tc.cidr),
					},
				}
				err := config.Validate()
				if tc.expectErr {
					g.Expect(err).To(HaveOccurred())
				} else {
					g.Expect(err).To(BeNil())
				}
			})
		})
	}
}

func TestValidateExternalServers(t *testing.T) {
	for _, tc := range []struct {
		name          string
		clusterConfig types.ClusterConfig
		expectErr     bool
	}{
		{name: "EmptyExternalServers", clusterConfig: types.ClusterConfig{Datastore: types.Datastore{ExternalServers: nil}}},
		{
			name: "ValidSingleExternalServers", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: vals.Pointer([]string{"localhost:123"}),
				},
			},
		},
		{
			name: "ValidMultipleExternalServers", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: vals.Pointer([]string{"https://localhost:123", "10.11.12.13:1234"}),
				},
			},
		},
		{
			name: "InvalidSingleExternalServers", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: vals.Pointer([]string{"localhost"}),
				},
			},
			expectErr: true,
		},
		{
			name: "InvalidMultipleExternalServers", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: vals.Pointer([]string{"localhost:123", "invalid_address:1:2"}),
				},
			},
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			tc.clusterConfig.SetDefaults()

			err := tc.clusterConfig.Validate()
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}
