package pki_test

import (
	"os"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestEtcdPKI(t *testing.T) {
	c := pki.NewEtcdPKI(pki.EtcdPKIOpts{
		Hostname:          "test",
		Years:             10,
		AllowSelfSignedCA: true,
	})

	g := NewWithT(t)
	g.Expect(c.CompleteCertificates()).To(Succeed())

	_, err := setup.EnsureEtcdPKI(&mock.Snap{
		Mock: mock.Mock{
			UID:              os.Getuid(),
			GID:              os.Getgid(),
			EtcdPKIDir:       "testdata",
			KubernetesPKIDir: "testdata",
		},
	}, c)
	g.Expect(err).To(BeNil())
}
