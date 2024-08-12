package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
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
						PodCIDR:     utils.Pointer(tc.cidr),
						ServiceCIDR: utils.Pointer("10.1.0.0/16"),
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
						PodCIDR:     utils.Pointer("10.1.0.0/16"),
						ServiceCIDR: utils.Pointer(tc.cidr),
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
		{name: "Empty", clusterConfig: types.ClusterConfig{Datastore: types.Datastore{ExternalServers: nil}}},
		{
			name: "HostPort", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"localhost:123"}),
				},
			},
		},
		{
			name: "FQDN", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"172.22.1.1.ec2.internal"}),
				},
			},
		},
		{
			name: "IPv4", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"10.11.12.13"}),
				},
			},
		},
		{
			name: "IPv6", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]"}),
				},
			},
		},
		{
			name: "ValidMultiple", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"https://localhost:123", "10.11.12.13"}),
				},
			},
		},
		{
			name: "InvalidSingle", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"invalid_address:1:2"}),
				},
			},
			expectErr: true,
		},
		{
			name: "InvalidMultiple", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"localhost:123", "invalid_address:1:2"}),
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
