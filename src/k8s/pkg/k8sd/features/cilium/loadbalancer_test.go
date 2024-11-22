package cilium_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/client/helm"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

// NOTE(hue): status.Message is not checked sometimes to avoid unnecessary complexity

func TestLoadBalancerDisabled(t *testing.T) {
	t.Run("HelmApplyFails", func(t *testing.T) {
		g := NewWithT(t)

		applyErr := errors.New("failed to apply")
		helmM := &helmmock.Mock{
			ApplyErr: applyErr,
		}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		lbCfg := types.LoadBalancer{
			Enabled: ptr.To(false),
		}

		status, err := cilium.ApplyLoadBalancer(context.Background(), snapM, lbCfg, types.Network{}, nil)

		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.LbDeleteFailedMsgTmpl, err)))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium.ChartCiliumLoadBalancer))
		g.Expect(callArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(callArgs.Values).To(BeNil())
	})

	t.Run("Success", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		lbCfg := types.LoadBalancer{
			Enabled: ptr.To(false),
		}
		networkCfg := types.Network{
			Enabled: ptr.To(true),
		}

		status, err := cilium.ApplyLoadBalancer(context.Background(), snapM, lbCfg, networkCfg, nil)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(cilium.DisabledMsg))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(2))

		firstCallArgs := helmM.ApplyCalledWith[0]
		g.Expect(firstCallArgs.Chart).To(Equal(cilium.ChartCiliumLoadBalancer))
		g.Expect(firstCallArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(firstCallArgs.Values).To(BeNil())

		// checking helm apply for network since it's enabled
		secondCallArgs := helmM.ApplyCalledWith[1]
		g.Expect(secondCallArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(secondCallArgs.State).To(Equal(helm.StateUpgradeOnlyOrDeleted(networkCfg.GetEnabled())))
	})
}

func TestLoadBalancerEnabled(t *testing.T) {
	t.Run("HelmApplyFails", func(t *testing.T) {
		g := NewWithT(t)

		applyErr := errors.New("failed to apply")
		helmM := &helmmock.Mock{
			ApplyErr: applyErr,
		}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		lbCfg := types.LoadBalancer{
			Enabled: ptr.To(true),
			// setting both modes to true for testing purposes
			L2Mode:  ptr.To(true),
			BGPMode: ptr.To(true),
		}
		networkCfg := types.Network{
			Enabled: ptr.To(true),
		}

		status, err := cilium.ApplyLoadBalancer(context.Background(), snapM, lbCfg, networkCfg, nil)

		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.LbDeployFailedMsgTmpl, err)))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StateUpgradeOnlyOrDeleted(networkCfg.GetEnabled())))
		l2announcements, ok := callArgs.Values["l2announcements"].(map[string]any)
		g.Expect(ok).To(BeTrue())
		g.Expect(l2announcements["enabled"]).To(Equal(lbCfg.GetL2Mode()))
		bgpControlPlane, ok := callArgs.Values["bgpControlPlane"].(map[string]any)
		g.Expect(ok).To(BeTrue())
		g.Expect(bgpControlPlane["enabled"]).To(Equal(lbCfg.GetBGPMode()))
	})

	for _, tc := range []struct {
		name          string
		l2Mode        bool
		bGPMode       bool
		statusMessage string
	}{
		{
			name:          "SuccessL2Mode",
			l2Mode:        true,
			bGPMode:       false,
			statusMessage: "enabled, L2 mode",
		},
		{
			name:          "SuccessBGPMode",
			l2Mode:        false,
			bGPMode:       true,
			statusMessage: "enabled, BGP mode",
		},
		{
			name:          "SuccessUnknownMode",
			l2Mode:        false,
			bGPMode:       false,
			statusMessage: "enabled, Unknown mode",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
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
			fd, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
			g.Expect(ok).To(BeTrue())
			fd.Resources = []*metav1.APIResourceList{
				{
					GroupVersion: "cilium.io/v2alpha1",
					APIResources: []metav1.APIResource{
						{Name: "ciliuml2announcementpolicies"},
						{Name: "ciliumloadbalancerippools"},
						{Name: "ciliumbgppeeringpolicies"},
					},
				},
			}
			snapM := &snapmock.Snap{
				Mock: snapmock.Mock{
					HelmClient:       helmM,
					KubernetesClient: &kubernetes.Client{Interface: clientset},
				},
			}
			lbCfg := types.LoadBalancer{
				Enabled: ptr.To(true),
				// setting both modes to true for testing purposes
				L2Mode:         ptr.To(tc.l2Mode),
				L2Interfaces:   ptr.To([]string{"eth0", "eth1"}),
				BGPMode:        ptr.To(tc.bGPMode),
				BGPLocalASN:    ptr.To(64512),
				BGPPeerAddress: ptr.To("10.0.0.1/32"),
				BGPPeerASN:     ptr.To(64513),
				BGPPeerPort:    ptr.To(179),
				CIDRs:          ptr.To([]string{"192.0.2.0/24"}),
				IPRanges: ptr.To([]types.LoadBalancer_IPRange{
					{Start: "20.0.20.100", Stop: "20.0.20.200"},
				}),
			}
			networkCfg := types.Network{
				Enabled: ptr.To(true),
			}

			status, err := cilium.ApplyLoadBalancer(context.Background(), snapM, lbCfg, networkCfg, nil)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.Enabled).To(BeTrue())
			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(status.Message).To(Equal(tc.statusMessage))

			g.Expect(helmM.ApplyCalledWith).To(HaveLen(2))

			firstCallArgs := helmM.ApplyCalledWith[0]
			g.Expect(firstCallArgs.Chart).To(Equal(cilium.ChartCilium))
			g.Expect(firstCallArgs.State).To(Equal(helm.StateUpgradeOnlyOrDeleted(networkCfg.GetEnabled())))
			l2announcements, ok := firstCallArgs.Values["l2announcements"].(map[string]any)
			g.Expect(ok).To(BeTrue())
			g.Expect(l2announcements["enabled"]).To(Equal(lbCfg.GetL2Mode()))
			bgpControlPlane, ok := firstCallArgs.Values["bgpControlPlane"].(map[string]any)
			g.Expect(ok).To(BeTrue())
			g.Expect(bgpControlPlane["enabled"]).To(Equal(lbCfg.GetBGPMode()))

			secondCallArgs := helmM.ApplyCalledWith[1]
			g.Expect(secondCallArgs.Chart).To(Equal(cilium.ChartCiliumLoadBalancer))
			g.Expect(secondCallArgs.State).To(Equal(helm.StatePresent))
			validateLoadBalancerValues(t, secondCallArgs.Values, lbCfg)

			// check if cilium-operator and cilium daemonset are restarted
			deployment, err := clientset.AppsV1().Deployments("kube-system").Get(context.Background(), "cilium-operator", metav1.GetOptions{})
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(deployment.Spec.Template.Annotations).To(HaveKey("kubectl.kubernetes.io/restartedAt"))
			daemonSet, err := clientset.AppsV1().DaemonSets("kube-system").Get(context.Background(), "cilium", metav1.GetOptions{})
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(daemonSet.Spec.Template.Annotations).To(HaveKey("kubectl.kubernetes.io/restartedAt"))
		})
	}
}

func validateLoadBalancerValues(t *testing.T, values map[string]interface{}, lbCfg types.LoadBalancer) {
	g := NewWithT(t)

	l2, ok := values["l2"].(map[string]any)
	g.Expect(ok).To(BeTrue())
	g.Expect(l2["enabled"]).To(Equal(lbCfg.GetL2Mode()))
	g.Expect(l2["interfaces"]).To(Equal(lbCfg.GetL2Interfaces()))

	ipPool, ok := values["ipPool"].(map[string]any)
	g.Expect(ok).To(BeTrue())
	cidrs, ok := ipPool["cidrs"].([]map[string]any)
	g.Expect(ok).To(BeTrue())
	g.Expect(cidrs).To(HaveLen(len(lbCfg.GetIPRanges()) + len(lbCfg.GetCIDRs())))
	for _, cidr := range lbCfg.GetCIDRs() {
		g.Expect(cidrs).To(ContainElement(map[string]any{"cidr": cidr}))
	}
	for _, ipRange := range lbCfg.GetIPRanges() {
		g.Expect(cidrs).To(ContainElement(map[string]any{"start": ipRange.Start, "stop": ipRange.Stop}))
	}

	bgp, ok := values["bgp"].(map[string]any)
	g.Expect(ok).To(BeTrue())
	g.Expect(bgp["enabled"]).To(Equal(lbCfg.GetBGPMode()))
	g.Expect(bgp["localASN"]).To(Equal(lbCfg.GetBGPLocalASN()))
	neighbors, ok := bgp["neighbors"].([]map[string]any)
	g.Expect(ok).To(BeTrue())
	g.Expect(neighbors).To(HaveLen(1))
	g.Expect(neighbors[0]["peerAddress"]).To(Equal(lbCfg.GetBGPPeerAddress()))
	g.Expect(neighbors[0]["peerASN"]).To(Equal(lbCfg.GetBGPPeerASN()))
	g.Expect(neighbors[0]["peerPort"]).To(Equal(lbCfg.GetBGPPeerPort()))
}
