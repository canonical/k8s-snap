package pki_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
)

func TestExternalDatastorePKI_CheckCertificates(t *testing.T) {
	tests := []struct {
		name          string
		pki           pki.ExternalDatastorePKI
		expectedError bool
	}{
		{
			name: "CheckCertificates with missing client certificate",
			pki: pki.ExternalDatastorePKI{
				DatastoreClientKey: "datastoreClientKey",
			},
			expectedError: true,
		},
		{
			name: "CheckCertificates with missing client key",
			pki: pki.ExternalDatastorePKI{
				DatastoreClientCert: "datastoreClientCert",
			},
			expectedError: true,
		},
		{
			name: "CheckCertificates with both client certificate and key",
			pki: pki.ExternalDatastorePKI{
				DatastoreClientCert: "datastoreClientCert",
				DatastoreClientKey:  "datastoreClientKey",
			},
			expectedError: false,
		},
		{
			name:          "CheckCertificates with neither client certificate nor key",
			pki:           pki.ExternalDatastorePKI{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pki.CheckCertificates()

			if (err != nil) != tt.expectedError {
				t.Errorf("Unexpected error status. Expected error: %v, got error: %v", tt.expectedError, err)
			}
		})
	}
}
