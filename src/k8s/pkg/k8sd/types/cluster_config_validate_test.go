package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestValidateCIDR(t *testing.T) {
	for _, tc := range []struct {
		cidr         string
		expectPodErr bool
		expectSvcErr bool
	}{
		{cidr: "192.168.0.0/16"},
		{cidr: "2001:0db8::/108"},
		{cidr: "10.2.0.0/16,2001:0db8::/108"},
		{cidr: "", expectPodErr: true, expectSvcErr: true},
		{cidr: "bananas", expectPodErr: true, expectSvcErr: true},
		{cidr: "fd01::/108,fd02::/108,fd03::/108", expectPodErr: true, expectSvcErr: true},
		{cidr: "10.1.0.0/32", expectPodErr: true, expectSvcErr: true},
		{cidr: "2001:0db8::/32", expectSvcErr: true},
	} {
		t.Run(tc.cidr, func(t *testing.T) {
			t.Run("Pod", func(t *testing.T) {
				g := NewWithT(t)
				config := types.ClusterConfig{
					Network: types.Network{
						PodCIDR:     utils.Pointer(tc.cidr),
						ServiceCIDR: utils.Pointer("10.1.0.0/16"),
					},
				}
				err := config.Validate()
				if tc.expectPodErr {
					g.Expect(err).To(HaveOccurred())
				} else {
					g.Expect(err).To(Not(HaveOccurred()))
				}
			})
			t.Run("Service", func(t *testing.T) {
				g := NewWithT(t)
				config := types.ClusterConfig{
					Network: types.Network{
						PodCIDR:     utils.Pointer("10.1.0.0/16"),
						ServiceCIDR: utils.Pointer(tc.cidr),
					},
				}
				err := config.Validate()
				if tc.expectSvcErr {
					g.Expect(err).To(HaveOccurred())
				} else {
					g.Expect(err).To(Not(HaveOccurred()))
				}
			})
		})
	}
}

func TestValidateExternalServers(t *testing.T) {
	for _, tc := range []struct {
		name          string
		clusterConfig types.ClusterConfig
		expectErr     bool
	}{
		{name: "Empty", clusterConfig: types.ClusterConfig{Datastore: types.Datastore{ExternalServers: nil}}},
		{
			name: "HostPort", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"localhost:123"}),
				},
			},
		},
		{
			name: "FQDN", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"172.22.1.1.ec2.internal"}),
				},
			},
		},
		{
			name: "IPv4", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"10.11.12.13"}),
				},
			},
		},
		{
			name: "IPv6", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]"}),
				},
			},
		},
		{
			name: "ValidMultiple", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"https://localhost:123", "10.11.12.13"}),
				},
			},
		},
		{
			name: "InvalidSingle", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"invalid_address:1:2"}),
				},
			},
			expectErr: true,
		},
		{
			name: "InvalidMultiple", clusterConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					ExternalServers: utils.Pointer([]string{"localhost:123", "invalid_address:1:2"}),
				},
			},
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			tc.clusterConfig.SetDefaults()

			err := tc.clusterConfig.Validate()
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(Not(HaveOccurred()))
			}
		})
	}
}

func TestValidateControlPlaneEndpoint(t *testing.T) {
	for _, tc := range []struct {
		name      string
		endpoint  types.ControlPlaneEndpoint
		expectErr bool
	}{
		{name: "NoEndpoint"},
		{
			name:     "External/IPv4",
			endpoint: types.ControlPlaneEndpoint{Host: utils.Pointer("10.0.0.250"), Backend: utils.Pointer("external")},
		},
		{
			name:     "External/IPv6",
			endpoint: types.ControlPlaneEndpoint{Host: utils.Pointer("2001:db8::1"), Backend: utils.Pointer("external")},
		},
		{
			name:     "Service/DNS",
			endpoint: types.ControlPlaneEndpoint{Host: utils.Pointer("api.example.com"), Backend: utils.Pointer("service")},
		},
		{
			name:      "InvalidHost",
			endpoint:  types.ControlPlaneEndpoint{Host: utils.Pointer("not a valid host!")},
			expectErr: true,
		},
		{
			name:      "InvalidBackend",
			endpoint:  types.ControlPlaneEndpoint{Host: utils.Pointer("10.0.0.250"), Backend: utils.Pointer("kube-vip")},
			expectErr: true,
		},
		{
			name:      "PortOutOfRange",
			endpoint:  types.ControlPlaneEndpoint{Host: utils.Pointer("10.0.0.250"), Port: utils.Pointer(70000), Backend: utils.Pointer("external")},
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			config := types.ClusterConfig{ControlPlaneEndpoint: tc.endpoint}
			config.SetDefaults()

			err := config.Validate()
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(Not(HaveOccurred()))
			}
		})
	}
}

func TestControlPlaneEndpointSANs(t *testing.T) {
	for _, tc := range []struct {
		name      string
		host      string
		expectIPs []string
		expectDNS []string
	}{
		{name: "Empty"},
		{name: "IPv4", host: "10.0.0.250", expectIPs: []string{"10.0.0.250"}},
		{name: "IPv6", host: "2001:db8::1", expectIPs: []string{"2001:db8::1"}},
		{name: "DNS", host: "api.example.com", expectDNS: []string{"api.example.com"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			ep := types.ControlPlaneEndpoint{}
			if tc.host != "" {
				ep.Host = utils.Pointer(tc.host)
			}
			ips, dnsNames := ep.SANs()

			gotIPs := make([]string, 0, len(ips))
			for _, ip := range ips {
				gotIPs = append(gotIPs, ip.String())
			}
			if len(tc.expectIPs) == 0 {
				g.Expect(gotIPs).To(BeEmpty())
			} else {
				g.Expect(gotIPs).To(Equal(tc.expectIPs))
			}
			if len(tc.expectDNS) == 0 {
				g.Expect(dnsNames).To(BeEmpty())
			} else {
				g.Expect(dnsNames).To(Equal(tc.expectDNS))
			}
		})
	}
}
