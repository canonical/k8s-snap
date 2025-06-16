package pki_test

import (
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	. "github.com/onsi/gomega"
)

func TestEtcdPKI(t *testing.T) {
	notBefore := time.Now()
	c := pki.NewEtcdPKI(pki.EtcdPKIOpts{
		Hostname:          "test",
		NotBefore:         notBefore,
		NotAfter:          notBefore.AddDate(1, 0, 0),
		AllowSelfSignedCA: true,
	})

	g := NewWithT(t)
	g.Expect(c.CompleteCertificates()).To(Succeed())
}
