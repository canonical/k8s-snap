package pki

import (
	"crypto/x509/pkix"
	"net"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestControlPlanePKI_CompleteWorkerNodePKI(t *testing.T) {

	g := NewWithT(t)
	serverCACert, serverCAKey, err := generateSelfSignedCA(pkix.Name{CommonName: "kubernetes-ca"}, 1, 2048)
	g.Expect(err).ToNot(HaveOccurred())
	clientCACert, clientCAKey, err := generateSelfSignedCA(pkix.Name{CommonName: "kubernetes-ca-client"}, 1, 2048)
	g.Expect(err).ToNot(HaveOccurred())

	for _, tc := range []struct {
		name        string
		withCerts   func(*ControlPlanePKI)
		expectErr   bool
		expectPKITo types.GomegaMatcher
	}{
		{
			name: "WithCACertAndKeys",
			withCerts: func(pki *ControlPlanePKI) {
				pki.CACert = serverCACert
				pki.CAKey = serverCAKey
				pki.ClientCACert = clientCACert
				pki.ClientCAKey = clientCAKey
			},
			expectPKITo: SatisfyAll(
				HaveField("CACert", Equal(serverCACert)),
				HaveField("KubeletCert", Not(BeEmpty())),
				HaveField("KubeletKey", Not(BeEmpty())),
				HaveField("KubeletClientCert", Not(BeEmpty())),
				HaveField("KubeletClientKey", Not(BeEmpty())),
				HaveField("KubeProxyClientCert", Not(BeEmpty())),
				HaveField("KubeProxyClientKey", Not(BeEmpty())),
			),
		},
		{
			name:      "WithoutCerts",
			withCerts: func(pki *ControlPlanePKI) {},
			expectErr: true,
		},
		{
			name: "WithoutCACert",
			withCerts: func(pki *ControlPlanePKI) {
				pki.ClientCACert = clientCACert
			},
			expectErr: true,
		},
		{
			name: "WithoutClientCACert",
			withCerts: func(pki *ControlPlanePKI) {
				pki.CACert = serverCACert
			},
			expectErr: true,
		},
		{
			name: "OnlyServerCAKey",
			withCerts: func(pki *ControlPlanePKI) {
				pki.CACert = serverCACert
				pki.CAKey = serverCAKey
				pki.ClientCACert = clientCACert
			},
			expectErr: true,
		},
		{
			name: "OnlyClientCAKey",
			withCerts: func(pki *ControlPlanePKI) {
				pki.CACert = serverCACert
				pki.ClientCACert = clientCACert
				pki.ClientCAKey = clientCAKey
			},
			expectPKITo: SatisfyAll(
				HaveField("CACert", Equal(serverCACert)),
				HaveField("KubeletCert", BeEmpty()),
				HaveField("KubeletKey", BeEmpty()),
				HaveField("KubeletClientCert", Not(BeEmpty())),
				HaveField("KubeletClientKey", Not(BeEmpty())),
				HaveField("KubeProxyClientCert", Not(BeEmpty())),
				HaveField("KubeProxyClientKey", Not(BeEmpty())),
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			cp := NewControlPlanePKI(ControlPlanePKIOpts{Years: 10})
			tc.withCerts(cp)

			pki, err := cp.CompleteWorkerNodePKI("worker", net.IP{10, 0, 0, 1}, 2048)
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(pki).To(tc.expectPKITo)
			}
		})
	}
}
