package pki

import (
	"crypto/x509/pkix"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_generateSelfSignedCA(t *testing.T) {
	cert, key, err := generateSelfSignedCA(pkix.Name{CommonName: "test-cert"}, 10, 2048)

	g := NewWithT(t)
	g.Expect(err).To(BeNil())
	g.Expect(cert).ToNot(BeEmpty())
	g.Expect(key).ToNot(BeEmpty())

	t.Run("Load", func(t *testing.T) {
		c, k, err := loadCertificate(cert, key)
		g.Expect(err).To(BeNil())
		g.Expect(c).ToNot(BeNil())
		g.Expect(k).ToNot(BeNil())
	})

	t.Run("LoadKeyOnly", func(t *testing.T) {
		cert, key, err := loadCertificate(cert, "")
		g.Expect(err).To(BeNil())
		g.Expect(cert).ToNot(BeNil())
		g.Expect(key).To(BeNil())
	})
}
