package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/shared"
	microclusterTypes "github.com/canonical/microcluster/v3/rest/types"
	"github.com/canonical/microcluster/v3/state"
)

// onPreInit is called before we bootstrap or join a node.
func (a *App) onPreInit(ctx context.Context, s state.State, bootstrap bool, initConfig map[string]string) error {
	var extraSANs []string
	if bootstrap {
		bootstrapConfig, err := utils.MicroclusterBootstrapConfigFromMap(initConfig)
		if err != nil {
			return fmt.Errorf("failed to get bootstrap config: %w", err)
		}
		extraSANs = bootstrapConfig.ExtraSANs
	} else {
		controlPlaneJoinConfig, err := utils.MicroclusterControlPlaneJoinConfigFromMap(initConfig)
		if err != nil {
			return fmt.Errorf("failed to get control plane join config: %w", err)
		}
		extraSANs = controlPlaneJoinConfig.ExtraSANS
	}

	err := os.Remove(filepath.Join(s.FileSystem().StateDir, "server.crt"))
	if err != nil {
		return fmt.Errorf("failed to remove server.crt: %w", err)
	}

	err = os.Remove(filepath.Join(s.FileSystem().StateDir, "server.key"))
	if err != nil {
		return fmt.Errorf("failed to remove server.key: %w", err)
	}

	cert, err := shared.KeyPairAndCA(
		s.FileSystem().StateDir,
		string(microclusterTypes.ServerCertificateName),
		shared.CertServer,
		shared.CertOptions{
			AddHosts:                true,
			CommonName:              s.Name(),
			SubjectAlternativeNames: append(extraSANs, s.Name()),
		})
	if err != nil {
		return err
	}

	err = a.client.UpdateCertificate(ctx, microclusterTypes.ServerCertificateName, microclusterTypes.KeyPair{
		Cert: string(cert.PublicKey()),
		Key:  string(cert.PrivateKey()),
	})
	if err != nil {
		return fmt.Errorf("failed to update certificate %s: %w", microclusterTypes.ServerCertificateName, err)
	}

	return nil
}
