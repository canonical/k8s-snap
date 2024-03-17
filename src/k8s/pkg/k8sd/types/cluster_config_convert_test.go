package types_test

import (
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestClusterConfigFromBootstrapConfig(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		g := NewWithT(t)

		bootstrapConfig := &apiv1.BootstrapConfig{
			ClusterCIDR:   "10.1.0.0/16",
			ServiceCIDR:   "10.152.183.0/24",
			Components:    []string{"dns", "network"},
			EnableRBAC:    vals.Pointer(true),
			K8sDqlitePort: 12345,
		}

		expectedConfig := types.ClusterConfig{
			APIServer: types.APIServer{
				AuthorizationMode: vals.Pointer("Node,RBAC"),
			},
			Datastore: types.Datastore{
				Type:          vals.Pointer("k8s-dqlite"),
				K8sDqlitePort: vals.Pointer(12345),
			},
			Network: types.Network{
				Enabled:     vals.Pointer(true),
				PodCIDR:     vals.Pointer("10.1.0.0/16"),
				ServiceCIDR: vals.Pointer("10.152.183.0/24"),
			},
			DNS: types.DNS{
				Enabled: vals.Pointer(true),
			},
		}

		g.Expect(types.ClusterConfigFromBootstrapConfig(bootstrapConfig)).To(Equal(expectedConfig))
	})

	t.Run("RBAC", func(t *testing.T) {
		for _, tc := range []struct {
			name                      string
			enableRBAC                *bool
			expectedAuthorizationMode *string
		}{
			{name: "EnableRBAC=true", enableRBAC: vals.Pointer(true), expectedAuthorizationMode: vals.Pointer("Node,RBAC")},
			{name: "EnableRBAC=false", enableRBAC: vals.Pointer(false), expectedAuthorizationMode: vals.Pointer("AlwaysAllow")},
			{name: "EnableRBAC=nil", enableRBAC: nil, expectedAuthorizationMode: vals.Pointer("Node,RBAC")},
		} {

			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)
				c := types.ClusterConfigFromBootstrapConfig(&apiv1.BootstrapConfig{EnableRBAC: tc.enableRBAC})
				g.Expect(c.APIServer.AuthorizationMode).To(Equal(tc.expectedAuthorizationMode))
			})
		}
	})
}
