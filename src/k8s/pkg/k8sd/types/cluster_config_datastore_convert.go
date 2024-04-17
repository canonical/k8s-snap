package types

import (
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
)

// DatastoreConfigFromUserFacing converts UserFacingDatastoreConfig from public API into a Datastore config.
func DatastoreConfigFromUserFacing(u apiv1.UserFacingDatastoreConfig) (Datastore, error) {
	// Changing the datastore configuration is opt-in. We expect the caller to explicitly set the "external" type.
	// The nil check is required to ensure we only fail if the DatastoreConfig is expected to change.
	if u.Type != nil && u.GetType() != "external" {
		return Datastore{}, fmt.Errorf("failed to updated datastore config: type must be %q but is %q", "external", u.GetType())
	}

	return Datastore{
		Type:               u.Type,
		ExternalServers:    u.Servers,
		ExternalCACert:     u.CACert,
		ExternalClientCert: u.CACert,
		ExternalClientKey:  u.ClientKey,
	}, nil
}

// ToUserFacing converts a Datastore to a UserFacingDatastoreConfig from the public API.
func (c Datastore) ToUserFacing() apiv1.UserFacingDatastoreConfig {
	return apiv1.UserFacingDatastoreConfig{
		Type:       c.Type,
		Servers:    c.ExternalServers,
		CACert:     c.ExternalCACert,
		ClientCert: c.ExternalCACert,
		ClientKey:  c.ExternalClientKey,
	}
}
