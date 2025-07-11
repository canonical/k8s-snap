package pki

import (
	"net"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestEtcdPKI_CompleteCertificates(t *testing.T) {
	notBefore := time.Now()

	g := NewWithT(t)

	testPki := &EtcdPKI{
		allowSelfSignedCA: true,
		hostname:          "localhost",
		notBefore:         notBefore,
		notAfter:          notBefore.AddDate(1, 0, 0),
		ipSANs:            []net.IP{net.ParseIP("127.0.0.1")},
		dnsSANs:           []string{"localhost"},
	}
	err := testPki.CompleteCertificates()
	g.Expect(err).ToNot(HaveOccurred())

	tests := []struct {
		name          string
		pki           *EtcdPKI
		expectedError bool
	}{
		{
			name:          "CompleteCertificates with missing ServerCert and ServerKey self-signing not allowed",
			pki:           &EtcdPKI{},
			expectedError: true,
		},
		{
			name: "CompleteCertificates with ServerCert but missing ServerKey",
			pki: &EtcdPKI{
				ServerCert: testPki.ServerCert,
			},
			expectedError: true,
		},
		{
			name: "CompleteCertificates with missing ServerCert but ServerKey",
			pki: &EtcdPKI{
				ServerKey: testPki.ServerKey,
			},
			expectedError: true,
		},
		{
			name: "CompleteCertificates with external CA",
			pki: &EtcdPKI{
				CACert:              testPki.CACert,
				ServerCert:          testPki.ServerCert,
				ServerKey:           testPki.ServerKey,
				ServerPeerCert:      testPki.ServerPeerCert,
				ServerPeerKey:       testPki.ServerPeerKey,
				APIServerClientCert: testPki.APIServerClientCert,
				APIServerClientKey:  testPki.APIServerClientKey,
			},
			expectedError: false,
		},
		{
			name: "CompleteCertificates with self-signed CA and successful certificate generation",
			pki: &EtcdPKI{
				allowSelfSignedCA: true,
				hostname:          "localhost",
				notBefore:         notBefore,
				notAfter:          notBefore.AddDate(1, 0, 0),
				ipSANs:            []net.IP{net.ParseIP("127.0.0.1")},
				dnsSANs:           []string{"localhost"},
			},
			expectedError: false,
		},
		{
			name: "CompleteCertificates with self-signed CA allowed",
			pki: &EtcdPKI{
				allowSelfSignedCA: true,
				hostname:          "localhost",
				notBefore:         notBefore,
				notAfter:          notBefore.AddDate(1, 0, 0),
				ipSANs:            []net.IP{},
				dnsSANs:           []string{"localhost"},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pki.CompleteCertificates()

			if tt.expectedError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestNewEtcdPKI(t *testing.T) {
	notBefore := time.Now()
	tests := []struct {
		name        string
		opts        EtcdPKIOpts
		expectedPki *EtcdPKI
	}{
		{
			name: "NewEtcdPKI with default values",
			opts: EtcdPKIOpts{
				Hostname:  "localhost",
				NotBefore: notBefore,
			},
			expectedPki: &EtcdPKI{
				hostname:  "localhost",
				notBefore: notBefore,
				notAfter:  notBefore.AddDate(1, 0, 0),
			},
		},
		{
			name: "NewEtcdPKI with custom values",
			opts: EtcdPKIOpts{
				Hostname:          "localhost",
				DNSSANs:           []string{"localhost"},
				IPSANs:            []net.IP{net.ParseIP("127.0.0.1")},
				NotBefore:         notBefore,
				NotAfter:          notBefore.AddDate(2, 0, 0),
				AllowSelfSignedCA: true,
			},
			expectedPki: &EtcdPKI{
				hostname:          "localhost",
				ipSANs:            []net.IP{net.ParseIP("127.0.0.1")},
				dnsSANs:           []string{"localhost"},
				notBefore:         notBefore,
				notAfter:          notBefore.AddDate(2, 0, 0),
				allowSelfSignedCA: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			pki := NewEtcdPKI(tt.opts)
			g.Expect(pki).To(Equal(tt.expectedPki), "Unexpected EtcdPKI")
		})
	}
}
