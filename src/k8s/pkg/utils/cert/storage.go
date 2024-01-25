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
func StoreCertKeyPair(cert string, key string, certPath string, keyPath string) error {
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

	logrus.WithField("cert_length", len(string(cert))).WithField("key_length", len(string(key))).Debugf("Writing %s cert and key to database", certName)
	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// TODO: this is a hack until we completely replace WriteCertKeyPairToK8sd()
		var clusterConfig database.ClusterConfig
		switch certName {
		case "certificates-ca":
			clusterConfig.Certificates.CACert = string(cert)
			clusterConfig.Certificates.CAKey = string(key)
		case "certificates-k8s-dqlite":
			clusterConfig.Certificates.K8sDqliteCert = string(cert)
			clusterConfig.Certificates.K8sDqliteKey = string(key)
		default:
			panic("only 'certificates-ca' or 'certificate-k8s-dqlite' is allowed")
		}
		if err := database.SetClusterConfig(ctx, tx, clusterConfig); err != nil {
			return fmt.Errorf("failed to set cluster config: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to write certificate %s to database: %w", certName, err)
	}
	return nil
}
