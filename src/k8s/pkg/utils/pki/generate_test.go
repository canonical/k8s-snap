package pkiutil_test

import (
	"crypto/x509/pkix"
	"testing"
	"time"

	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	. "github.com/onsi/gomega"
)

func TestGenerateSelfSignedCA(t *testing.T) {
	cert, key, err := pkiutil.GenerateSelfSignedCA(pkix.Name{CommonName: "test-cert"}, time.Now().AddDate(10, 0, 0), 2048)

	g := NewWithT(t)
	g.Expect(err).To(BeNil())
	g.Expect(cert).ToNot(BeEmpty())
	g.Expect(key).ToNot(BeEmpty())

	t.Run("Load", func(t *testing.T) {
		c, k, err := pkiutil.LoadCertificate(cert, key)
		g.Expect(err).To(BeNil())
		g.Expect(c).ToNot(BeNil())
		g.Expect(k).ToNot(BeNil())
	})

	t.Run("LoadCertOnly", func(t *testing.T) {
		cert, key, err := pkiutil.LoadCertificate(cert, "")
		g.Expect(err).To(BeNil())
		g.Expect(cert).ToNot(BeNil())
		g.Expect(key).To(BeNil())
	})
}
