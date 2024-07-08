package types_test

import (
	"github.com/canonical/k8s/pkg/utils"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
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
				Type:       utils.Pointer("external"),
				Servers:    utils.Pointer([]string{"server1", "server2"}),
				CACert:     utils.Pointer("ca_cert"),
				ClientCert: utils.Pointer("client_cert"),
				ClientKey:  utils.Pointer("client_key"),
			},
			expectedConfig: types.Datastore{
				Type:               utils.Pointer("external"),
				ExternalServers:    utils.Pointer([]string{"server1", "server2"}),
				ExternalCACert:     utils.Pointer("ca_cert"),
				ExternalClientCert: utils.Pointer("client_cert"),
				ExternalClientKey:  utils.Pointer("client_key"),
			},
		},
		{
			name: "Invalid datastore config type",
			userFacingConfig: apiv1.UserFacingDatastoreConfig{
				Type:       utils.Pointer("k8s-dqlite"),
				Servers:    utils.Pointer([]string{"server1", "server2"}),
				CACert:     utils.Pointer("ca_cert"),
				ClientCert: utils.Pointer("client_cert"),
				ClientKey:  utils.Pointer("client_key"),
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
				Type:               utils.Pointer("external"),
				ExternalServers:    utils.Pointer([]string{"server1", "server2"}),
				ExternalCACert:     utils.Pointer("ca_cert"),
				ExternalClientCert: utils.Pointer("client_cert"),
				ExternalClientKey:  utils.Pointer("client_key"),
			},
			expectedUserFacingConfig: apiv1.UserFacingDatastoreConfig{
				Type:       utils.Pointer("external"),
				Servers:    utils.Pointer([]string{"server1", "server2"}),
				CACert:     utils.Pointer("ca_cert"),
				ClientCert: utils.Pointer("client_cert"),
				ClientKey:  utils.Pointer("client_key"),
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
