package pki_test

import (
	"crypto/x509/pkix"
	"net"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestControlPlanePKI_CompleteWorkerNodePKI(t *testing.T) {

	g := NewWithT(t)
	serverCACert, serverCAKey, err := pkiutil.GenerateSelfSignedCA(pkix.Name{CommonName: "kubernetes-ca"}, time.Now().AddDate(1, 0, 0), 2048)
	g.Expect(err).ToNot(HaveOccurred())
	clientCACert, clientCAKey, err := pkiutil.GenerateSelfSignedCA(pkix.Name{CommonName: "kubernetes-ca-client"}, time.Now().AddDate(1, 0, 0), 2048)
	g.Expect(err).ToNot(HaveOccurred())

	for _, tc := range []struct {
		name        string
		withCerts   func(*pki.ControlPlanePKI)
		expectErr   bool
		expectPKITo types.GomegaMatcher
	}{
		{
			name: "WithCACertAndKeys",
			withCerts: func(pki *pki.ControlPlanePKI) {
				pki.CACert = serverCACert
				pki.CAKey = serverCAKey
				pki.ClientCACert = clientCACert
				pki.ClientCAKey = clientCAKey
			},
			expectPKITo: SatisfyAll(
				HaveField("CACert", Equal(serverCACert)),
				HaveField("ClientCACert", Equal(clientCACert)),
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
			withCerts: func(pki *pki.ControlPlanePKI) {},
			expectErr: true,
		},
		{
			name: "WithoutCACert",
			withCerts: func(pki *pki.ControlPlanePKI) {
				pki.ClientCACert = clientCACert
			},
			expectErr: true,
		},
		{
			name: "WithoutClientCACert",
			withCerts: func(pki *pki.ControlPlanePKI) {
				pki.CACert = serverCACert
			},
			expectErr: true,
		},
		{
			name: "OnlyServerCAKey",
			withCerts: func(pki *pki.ControlPlanePKI) {
				pki.CACert = serverCACert
				pki.CAKey = serverCAKey
				pki.ClientCACert = clientCACert
			},
			expectPKITo: SatisfyAll(
				HaveField("CACert", Equal(serverCACert)),
				HaveField("ClientCACert", Equal(clientCACert)),
				HaveField("KubeletCert", Not(BeEmpty())),
				HaveField("KubeletKey", Not(BeEmpty())),
				HaveField("KubeletClientCert", BeEmpty()),
				HaveField("KubeletClientKey", BeEmpty()),
				HaveField("KubeProxyClientCert", BeEmpty()),
				HaveField("KubeProxyClientKey", BeEmpty()),
			),
		},
		{
			name: "OnlyClientCAKey",
			withCerts: func(pki *pki.ControlPlanePKI) {
				pki.CACert = serverCACert
				pki.ClientCACert = clientCACert
				pki.ClientCAKey = clientCAKey
			},
			expectPKITo: SatisfyAll(
				HaveField("CACert", Equal(serverCACert)),
				HaveField("ClientCACert", Equal(clientCACert)),
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
			cp := pki.NewControlPlanePKI(pki.ControlPlanePKIOpts{ExpirationDate: time.Now().AddDate(1, 0, 0)})
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
