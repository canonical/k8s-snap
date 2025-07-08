package cilium_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/cilium"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	testenv "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

func TestReconcile(t *testing.T) {
	testenv.WithState(t, func(ctx context.Context, s state.State) {
		applyErr := errors.New("failed to apply")
		for _, tc := range []struct {
			name string

			// given
			apiServer    types.APIServer
			network      types.Network
			gateway      types.Gateway
			ingress      types.Ingress
			annotations  types.Annotations
			applyChanged bool
			helmErr      error

			// expected
			networkStatusEnabled bool
			networkStatusMsg     string
			gatewayStatusEnabled bool
			gatewayStatusMsg     string
			ingressStatusEnabled bool
			ingressStatusMsg     string
			expCiliumValues      map[string]any
		}{
			{
				name:      "HappyPath",
				apiServer: types.APIServer{SecurePort: ptr.To(5678)},
				network: types.Network{
					Enabled: ptr.To(true),
					PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
				},
				gateway: types.Gateway{Enabled: ptr.To(true)},
				ingress: types.Ingress{
					Enabled:             ptr.To(true),
					DefaultTLSSecret:    ptr.To("secret-name"),
					EnableProxyProtocol: ptr.To(true),
				},
				annotations: types.Annotations{
					apiv1_annotations.AnnotationDevices:             "eth0,eth1",
					apiv1_annotations.AnnotationDirectRoutingDevice: "eth0",
					apiv1_annotations.AnnotationVLANBPFBypass:       "100,200,300",
					apiv1_annotations.AnnotationCNIExclusive:        "true",
					apiv1_annotations.AnnotationSCTPEnabled:         "true",
					apiv1_annotations.AnnotationTunnelPort:          "8473",
				},

				networkStatusEnabled: true,
				networkStatusMsg:     cilium.EnabledMsg,
				gatewayStatusEnabled: true,
				gatewayStatusMsg:     cilium.EnabledMsg,
				ingressStatusEnabled: true,
				ingressStatusMsg:     cilium.EnabledMsg,
				expCiliumValues: map[string]any{
					"bpf": map[string]any{
						"vlanBypass": []int{100, 200, 300},
					},
					"image": map[string]any{
						"repository": cilium.CiliumAgentImageRepo,
						"tag":        cilium.CiliumAgentImageTag,
						"useDigest":  false,
					},
					"socketLB": map[string]any{
						"enabled": true,
					},
					"cni": map[string]any{
						"confPath":  "/etc/cni/net.d",
						"binPath":   "/opt/cni/bin",
						"exclusive": true,
					},
					"operator": map[string]any{
						"replicas": 1,
						"image": map[string]any{
							"repository": cilium.CiliumOperatorImageRepo,
							"tag":        cilium.CiliumOperatorImageTag,
							"useDigest":  false,
						},
					},
					"ipv4": map[string]any{
						"enabled": true,
					},
					"ipv6": map[string]any{
						"enabled": true,
					},
					"ipam": map[string]any{
						"operator": map[string]any{
							"clusterPoolIPv4PodCIDRList": "192.0.2.0/24",
							"clusterPoolIPv6PodCIDRList": "2001:db8::/32",
						},
					},
					"envoy": map[string]any{
						"enabled": false,
					},
					"nodePort": map[string]any{
						"enabled":             true,
						"enableHealthCheck":   false,
						"directRoutingDevice": "eth0",
					},
					"disableEnvoyVersionCheck": true,
					"k8sServiceHost":           "127.0.0.1", "k8sServicePort": 5678,
					"enableRuntimeDeviceDetection": true,
					"tunnelPort":                   8473,
					"devices":                      "eth0,eth1",
					"gatewayAPI": map[string]any{
						"enabled": true,
					},
					"ingressController": map[string]any{
						cilium.IngressOptionEnabled:                true,
						cilium.IngressOptionLoadBalancerMode:       cilium.IngressOptionLoadBalancerModeShared,
						cilium.IngressOptionDefaultSecretNamespace: cilium.IngressOptionDefaultSecretNamespaceKubeSystem,
						cilium.IngressOptionDefaultSecretName:      "secret-name",
						cilium.IngressOptionEnableProxyProtocol:    true,
					},
				},
			},
			{
				name: "DisableNetwork",
				network: types.Network{
					Enabled: ptr.To(false),
				},

				networkStatusEnabled: false,
				networkStatusMsg:     cilium.DisabledMsg,
				gatewayStatusEnabled: false,
				gatewayStatusMsg:     cilium.DisabledMsg,
				ingressStatusEnabled: false,
				ingressStatusMsg:     cilium.DisabledMsg,
			},
			{
				name:      "DisableGateway",
				apiServer: types.APIServer{SecurePort: ptr.To(5678)},
				network: types.Network{
					Enabled: ptr.To(true),
					PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
				},
				gateway: types.Gateway{Enabled: ptr.To(false)},
				ingress: types.Ingress{
					Enabled: ptr.To(true),
				},

				networkStatusEnabled: true,
				networkStatusMsg:     cilium.EnabledMsg,
				gatewayStatusEnabled: false,
				gatewayStatusMsg:     cilium.DisabledMsg,
				ingressStatusEnabled: true,
				ingressStatusMsg:     cilium.EnabledMsg,
			},
			{
				name:      "DisableIngress",
				apiServer: types.APIServer{SecurePort: ptr.To(5678)},
				network: types.Network{
					Enabled: ptr.To(true),
					PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
				},
				gateway: types.Gateway{Enabled: ptr.To(true)},
				ingress: types.Ingress{
					Enabled: ptr.To(false),
				},

				networkStatusEnabled: true,
				networkStatusMsg:     cilium.EnabledMsg,
				gatewayStatusEnabled: true,
				gatewayStatusMsg:     cilium.EnabledMsg,
				ingressStatusEnabled: false,
				ingressStatusMsg:     cilium.DisabledMsg,
			},
			{
				name:      "DisableGatewayAndIngress",
				apiServer: types.APIServer{SecurePort: ptr.To(5678)},
				network: types.Network{
					Enabled: ptr.To(true),
					PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
				},
				gateway: types.Gateway{Enabled: ptr.To(false)},
				ingress: types.Ingress{
					Enabled: ptr.To(false),
				},

				networkStatusEnabled: true,
				networkStatusMsg:     cilium.EnabledMsg,
				gatewayStatusEnabled: false,
				gatewayStatusMsg:     cilium.DisabledMsg,
				ingressStatusEnabled: false,
				ingressStatusMsg:     cilium.DisabledMsg,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)

				helmM := &helmmock.Mock{
					ApplyErr:     tc.helmErr,
					ApplyChanged: tc.applyChanged,
				}
				snapM := &snapmock.Snap{
					Mock: snapmock.Mock{
						HelmClient: helmM,
					},
				}

				statuses, err := cilium.ApplyCilium(context.Background(), snapM, s, tc.apiServer, tc.network, tc.gateway, tc.ingress, tc.annotations)

				if tc.helmErr == nil {
					g.Expect(err).To(Not(HaveOccurred()))
				} else {
					g.Expect(err).To(MatchError(applyErr))
				}

				g.Expect(statuses[features.Network].Enabled).To(Equal(tc.networkStatusEnabled))
				g.Expect(statuses[features.Network].Message).To(Equal(tc.networkStatusMsg))
				g.Expect(statuses[features.Network].Version).To(Equal(cilium.CiliumAgentImageTag))

				g.Expect(statuses[features.Gateway].Enabled).To(Equal(tc.gatewayStatusEnabled))
				g.Expect(statuses[features.Gateway].Message).To(Equal(tc.gatewayStatusMsg))
				g.Expect(statuses[features.Gateway].Version).To(Equal(cilium.CiliumAgentImageTag))

				g.Expect(statuses[features.Ingress].Enabled).To(Equal(tc.ingressStatusEnabled))
				g.Expect(statuses[features.Ingress].Message).To(Equal(tc.ingressStatusMsg))
				g.Expect(statuses[features.Ingress].Version).To(Equal(cilium.CiliumAgentImageTag))

				if tc.network.GetEnabled() {
					g.Expect(helmM.ApplyCalledWith).To(HaveLen(3))
					if tc.expCiliumValues != nil {
						g.Expect(mapsEqual(t, helmM.ApplyCalledWith[len(helmM.ApplyCalledWith)-1].Values, tc.expCiliumValues)).To(BeTrue())
					}
				} else {
					g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))
				}
			})
		}
	})
}

func mapsEqual(t *testing.T, m1, m2 map[string]any) bool {
	t.Helper()

	if len(m1) != len(m2) {
		return false
	}

	b1, err := json.MarshalIndent(m1, "", "  ")
	if err != nil {
		t.Error(err)
	}
	b2, err := json.MarshalIndent(m2, "", "  ")
	if err != nil {
		t.Error(err)
	}

	return bytes.Equal(b1, b2)
}
