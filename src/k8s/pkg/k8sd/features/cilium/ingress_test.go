package cilium_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
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
		networkEnabled bool
		applyChanged   bool
		ingressEnabled bool
		helmErr        error
		//then
		statusMsg     string
		statusEnabled bool
	}{
		{
			name:           "HelmFailNetworkEnabled",
			networkEnabled: true,
			helmErr:        applyErr,
			statusMsg:      fmt.Sprintf(cilium.IngressDeployFailedMsgTmpl, fmt.Errorf("failed to enable ingress: %w", applyErr)),
			statusEnabled:  false,
		},
		{
			name:           "HelmFailNetworkDisabled",
			networkEnabled: false,
			statusMsg:      fmt.Sprintf(cilium.IngressDeleteFailedMsgTmpl, fmt.Errorf("failed to disable ingress: %w", applyErr)),
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
				Enabled: ptr.To(tc.ingressEnabled),
			}

			status, err := cilium.ApplyIngress(context.Background(), snapM, ingress, network, nil)

			if tc.helmErr == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(applyErr))
			}
			g.Expect(status.Enabled).To(Equal(tc.statusEnabled))
			g.Expect(status.Message).To(Equal(tc.statusMsg))
			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
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

		status, err := cilium.ApplyIngress(context.Background(), snapM, ingress, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.IngressDeployFailedMsgTmpl, err)))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
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

		status, err := cilium.ApplyIngress(context.Background(), snapM, ingress, network, nil)

		g.Expect(err).To(BeNil())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Message).To(Equal(cilium.EnabledMsg))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
	})
}
