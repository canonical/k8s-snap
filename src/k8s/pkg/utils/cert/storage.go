package cert

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/microcluster/state"
	"github.com/sirupsen/logrus"
)

// StoreCertKeyPair read the certificate & key from the k8sd database and writes
// them to the specified path on disk.
func StoreCertKeyPair(ctx context.Context, state *state.State, certName string, certPath string, keyPath string) error {
	// Get the certificates from the k8sd cluster
	var cert, key string
	var err error
	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		cert, key, err = database.GetCertificateAndKey(ctx, tx, certName)
		if err != nil {
			return fmt.Errorf("failed to get %s certs and key from database: %w", certName, err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to perform %s certificate transaction request: %w", certName, err)
	}

	logrus.WithField("cert_length", len(string(cert))).WithField("key_length", len(string(key))).Debug("Writing k8s-dqlite cert and key to disk")
	if err := os.WriteFile(certPath, []byte(cert), 0644); err != nil {
		return fmt.Errorf("failed to write cert to %s: %w", certPath, err)
	}

	if err := os.WriteFile(keyPath, []byte(key), 0644); err != nil {
		return fmt.Errorf("failed to write key to %s: %w", keyPath, err)
	}
	return nil
}

// WriteCertKeyPairToK8sd gets local cert and key and stores them in the k8sd database.
func WriteCertKeyPairToK8sd(ctx context.Context, state *state.State, certName string, certPath string, keyPath string) error {
	// Read cert and key from local
	cert, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read cert from %s: %w", certPath, err)
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key from %s: %w", keyPath, err)
	}

	logrus.WithField("cert_length", len(string(cert))).WithField("key_length", len(string(key))).Debug("Writing k8s-dqlite cert and key to database")
	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err = database.CreateCertificate(ctx, tx, certName, string(cert), string(key))
		if err != nil {
			return fmt.Errorf("failed to write %s certs and key to database: %w", certName, err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to perform %s certificate transaction write request: %w", certName, err)
	}
	return nil
}
