package utils_test

import (
	"crypto/x509"
	"testing"

	"github.com/canonical/k8s/pkg/utils"

	. "github.com/onsi/gomega"
)

func TestSplitIPAndDNSSANs(t *testing.T) {
	tests := []string{"192.168.0.1", "::1", "cluster.local", "kubernetes.svc.local", "", "2001:db8:0:1:1:1:1:1"}

	g := NewWithT(t)
	gotIPs, gotDNSs := utils.SplitIPAndDNSSANs(tests)

	// Convert cert.IPAddresses to a slice of string representations
	ips := make([]string, len(gotIPs))
	for i, ip := range gotIPs {
		ips[i] = ip.String()
	}

	g.Expect(ips).To(ConsistOf("192.168.0.1", "::1", "2001:db8:0:1:1:1:1:1"))

	g.Expect(gotDNSs).To(ConsistOf("cluster.local", "kubernetes.svc.local"))
}

func TestTLSClientConfigWithTrustedCertificate(t *testing.T) {
	g := NewWithT(t)

	// Mock certificate and certificate pool for testing
	remoteCert := &x509.Certificate{
		DNSNames: []string{"bubblegum.com"},
	}
	rootCAs := x509.NewCertPool()

	tlsConfig, err := utils.TLSClientConfigWithTrustedCertificate(remoteCert, rootCAs)

	g.Expect(err).To(BeNil())
	g.Expect(tlsConfig.ServerName).To(Equal("bubblegum.com"))
	g.Expect(tlsConfig.RootCAs.Subjects()).To(ContainElement(remoteCert.RawSubject))

	// Test with invalid remote certificate
	tlsConfig, err = utils.TLSClientConfigWithTrustedCertificate(nil, rootCAs)
	g.Expect(err).ToNot(BeNil())
	g.Expect(tlsConfig).To(BeNil())

	// Test with nil root CAs
	_, err = utils.TLSClientConfigWithTrustedCertificate(remoteCert, nil)
	g.Expect(err).To(BeNil())
}
