package types

import apiv1 "github.com/canonical/k8s/api/v1"

// DatastoreConfigFromUserFacing converts UserFacingDatastoreConfig from public API into a Datastore config.
func DatastoreConfigFromUserFacing(u apiv1.UserFacingDatastoreConfig) Datastore {
	return Datastore{
		Type:               u.Type,
		ExternalURL:        u.Servers,
		ExternalCACert:     u.CACert,
		ExternalClientCert: u.CACert,
		ExternalClientKey:  u.ClientKey,
	}
}

// ToUserFacing converts a Datastore to a UserFacingDatastoreConfig from the public API.
func (c Datastore) ToUserFacing() apiv1.UserFacingDatastoreConfig {
	return apiv1.UserFacingDatastoreConfig{
		Type:       c.Type,
		Servers:    c.ExternalURL,
		CACert:     c.ExternalCACert,
		ClientCert: c.ExternalCACert,
		ClientKey:  c.ExternalClientKey,
	}
}
