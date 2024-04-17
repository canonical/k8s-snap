package types_test

import (
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestDatastoreConfigFromUserFacing(t *testing.T) {
	testCases := []struct {
		name             string
		userFacingConfig apiv1.UserFacingDatastoreConfig
		expectedConfig   types.Datastore
		expectedError    bool
	}{
		{
			name: "Valid external datastore config",
			userFacingConfig: apiv1.UserFacingDatastoreConfig{
				Type:       vals.Pointer("external"),
				Servers:    vals.Pointer([]string{"server1", "server2"}),
				CACert:     vals.Pointer("ca_cert"),
				ClientCert: vals.Pointer("client_cert"),
				ClientKey:  vals.Pointer("client_key"),
			},
			expectedConfig: types.Datastore{
				Type:               vals.Pointer("external"),
				ExternalServers:    vals.Pointer([]string{"server1", "server2"}),
				ExternalCACert:     vals.Pointer("ca_cert"),
				ExternalClientCert: vals.Pointer("client_cert"),
				ExternalClientKey:  vals.Pointer("client_key"),
			},
		},
		{
			name: "Invalid datastore config type",
			userFacingConfig: apiv1.UserFacingDatastoreConfig{
				Type:       vals.Pointer("k8s-dqlite"),
				Servers:    vals.Pointer([]string{"server1", "server2"}),
				CACert:     vals.Pointer("ca_cert"),
				ClientCert: vals.Pointer("client_cert"),
				ClientKey:  vals.Pointer("client_key"),
			},
			expectedConfig: types.Datastore{},
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			config, err := types.DatastoreConfigFromUserFacing(tc.userFacingConfig)

			if tc.expectedError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
			g.Expect(config).To(Equal(tc.expectedConfig))
		})
	}
}

func TestDatastoreToUserFacing(t *testing.T) {
	testCases := []struct {
		name                     string
		datastoreConfig          types.Datastore
		expectedUserFacingConfig apiv1.UserFacingDatastoreConfig
	}{
		{
			name: "Valid datastore to user-facing config",
			datastoreConfig: types.Datastore{
				Type:               vals.Pointer("external"),
				ExternalServers:    vals.Pointer([]string{"server1", "server2"}),
				ExternalCACert:     vals.Pointer("ca_cert"),
				ExternalClientCert: vals.Pointer("client_cert"),
				ExternalClientKey:  vals.Pointer("client_key"),
			},
			expectedUserFacingConfig: apiv1.UserFacingDatastoreConfig{
				Type:       vals.Pointer("external"),
				Servers:    vals.Pointer([]string{"server1", "server2"}),
				CACert:     vals.Pointer("ca_cert"),
				ClientCert: vals.Pointer("client_cert"),
				ClientKey:  vals.Pointer("client_key"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			userFacingConfig := tc.datastoreConfig.ToUserFacing()
			g.Expect(userFacingConfig).To(Equal(tc.expectedUserFacingConfig))
		})
	}
}
