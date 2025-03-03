package ingress_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/client/helm/loader"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	cilium_ingress "github.com/canonical/k8s/pkg/k8sd/features/cilium/ingress"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestIngress(t *testing.T) {
	applyErr := errors.New("failed to apply")
	for _, tc := range []struct {
		name string
		// given
		networkEnabled      bool
		applyChanged        bool
		ingressEnabled      bool
		defaultSecretName   string
		enableProxyProtocol bool
		helmErr             error
		// then
		statusMsg     string
		statusEnabled bool
	}{
		{
			name:           "HelmFailNetworkEnabled",
			networkEnabled: true,
			helmErr:        applyErr,
			statusMsg:      fmt.Sprintf(cilium_ingress.IngressDeployFailedMsgTmpl, fmt.Errorf("failed to enable ingress: %w", applyErr)),
			statusEnabled:  false,
		},
		{
			name:           "HelmFailNetworkDisabled",
			networkEnabled: false,
			statusMsg:      fmt.Sprintf(cilium_ingress.IngressDeleteFailedMsgTmpl, fmt.Errorf("failed to disable ingress: %w", applyErr)),
			statusEnabled:  false,
			helmErr:        applyErr,
		},
		{
			name:           "HelmUnchangedIngressEnabled",
			ingressEnabled: true,
			statusMsg:      cilium.EnabledMsg,
			statusEnabled:  true,
		},
		{
			name:           "HelmUnchangedIngressDisabled",
			ingressEnabled: false,
			statusMsg:      cilium.DisabledMsg,
			statusEnabled:  false,
		},
		{
			name:           "HelmChangedIngressDisabled",
			applyChanged:   true,
			ingressEnabled: false,
			statusMsg:      cilium.DisabledMsg,
			statusEnabled:  false,
		},
		{
			name:                "HelmUnchangedIngressEnabled/",
			ingressEnabled:      true,
			defaultSecretName:   "secret-name",
			enableProxyProtocol: true,
			statusMsg:           cilium.EnabledMsg,
			statusEnabled:       true,
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
			network := types.Network{
				Enabled: ptr.To(tc.networkEnabled),
			}
			ingress := types.Ingress{
				Enabled:             ptr.To(tc.ingressEnabled),
				DefaultTLSSecret:    ptr.To(tc.defaultSecretName),
				EnableProxyProtocol: ptr.To(tc.enableProxyProtocol),
			}
			mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))
			status, err := cilium_ingress.ApplyIngress(context.Background(), snapM, mc, ingress, network, nil)

			if tc.helmErr == nil {
				g.Expect(err).To(Not(HaveOccurred()))
			} else {
				g.Expect(err).To(MatchError(applyErr))
			}
			g.Expect(status.Enabled).To(Equal(tc.statusEnabled))
			g.Expect(status.Message).To(Equal(tc.statusMsg))
			g.Expect(status.Version).To(Equal(cilium_network.FeatureNetwork.GetImage(cilium_network.CiliumAgentImageName).Tag))
			g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

			callArgs := helmM.ApplyCalledWith[0]
			g.Expect(callArgs.Chart).To(Equal(cilium_network.FeatureNetwork.GetChart(cilium_network.CiliumChartName)))
			validateIngressValues(g, callArgs.Values, ingress)
		})
	}
}

func TestIngressRollout(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset(
			&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy",
					Namespace: "kube-system",
				},
			},
		)
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
				KubernetesClient: &kubernetes.Client{
					Interface: clientset,
				},
			},
		}
		network := types.Network{}
		ingress := types.Ingress{
			Enabled: ptr.To(true),
		}
		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))
		status, err := cilium_ingress.ApplyIngress(context.Background(), snapM, mc, ingress, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium_ingress.IngressDeployFailedMsgTmpl, err)))
		g.Expect(status.Version).To(Equal(cilium_network.FeatureNetwork.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium_network.FeatureNetwork.GetChart(cilium_network.CiliumChartName)))
		validateIngressValues(g, callArgs.Values, ingress)
	})

	t.Run("Success", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset(
			&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cilium-operator",
					Namespace: "kube-system",
				},
			},
			&v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cilium",
					Namespace: "kube-system",
				},
			},
		)
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
				KubernetesClient: &kubernetes.Client{
					Interface: clientset,
				},
			},
		}
		network := types.Network{}
		ingress := types.Ingress{
			Enabled: ptr.To(true),
		}
		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))
		status, err := cilium_ingress.ApplyIngress(context.Background(), snapM, mc, ingress, network, nil)

		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Message).To(Equal(cilium.EnabledMsg))
		g.Expect(status.Version).To(Equal(cilium_network.FeatureNetwork.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium_network.FeatureNetwork.GetChart(cilium_network.CiliumChartName)))
		validateIngressValues(g, callArgs.Values, ingress)
	})
}

func validateIngressValues(g Gomega, values map[string]any, ingress types.Ingress) {
	ingressController, ok := values["ingressController"].(map[string]any)
	g.Expect(ok).To(BeTrue())
	if ingress.GetEnabled() {
		g.Expect(ingressController[cilium_ingress.IngressOptionEnabled]).To(BeTrue())
		g.Expect(ingressController[cilium_ingress.IngressOptionLoadBalancerMode]).To(Equal(cilium_ingress.IngressOptionLoadBalancerModeShared))
		g.Expect(ingressController[cilium_ingress.IngressOptionDefaultSecretNamespace]).To(Equal(cilium_ingress.IngressOptionDefaultSecretNamespaceKubeSystem))
		g.Expect(ingressController[cilium_ingress.IngressOptionDefaultSecretName]).To(Equal(ingress.GetDefaultTLSSecret()))
		g.Expect(ingressController[cilium_ingress.IngressOptionEnableProxyProtocol]).To(Equal(ingress.GetEnableProxyProtocol()))
	} else {
		g.Expect(ingressController[cilium_ingress.IngressOptionEnabled]).To(BeFalse())
		g.Expect(ingressController[cilium_ingress.IngressOptionLoadBalancerMode]).To(Equal(""))
		g.Expect(ingressController[cilium_ingress.IngressOptionDefaultSecretNamespace]).To(Equal(""))
		g.Expect(ingressController[cilium_ingress.IngressOptionDefaultSecretName]).To(Equal(""))
		g.Expect(ingressController[cilium_ingress.IngressOptionEnableProxyProtocol]).To(BeFalse())
	}
}
