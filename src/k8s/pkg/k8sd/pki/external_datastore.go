package pki

import "fmt"

type ExternalDatastorePKI struct {
	DatastoreCACert, DatastoreClientCert, DatastoreClientKey string
}

// CheckCertificates checks missing or unset certificates.
func (c *ExternalDatastorePKI) CheckCertificates() error {
	// Fail hard if keys of self-signed certificates are set without the respective certificates
	switch {
	case c.DatastoreClientCert == "" && c.DatastoreClientKey != "":
		return fmt.Errorf("external datastore certificate key set without a certificate, fail to prevent further issues")
	case c.DatastoreClientCert != "" && c.DatastoreClientKey == "":
		return fmt.Errorf("external datastore certificate set without a key, fail to prevent further issues")
	}

	return nil
}
