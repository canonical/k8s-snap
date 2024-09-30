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
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestGatewayEnabled(t *testing.T) {
	t.Run("HelmApplyErr", func(t *testing.T) {
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
		network := types.Network{}
		gateway := types.Gateway{
			Enabled: ptr.To(true),
		}

		status, err := cilium.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.GatewayDeployFailedMsgTmpl, applyErr)))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

	})

	t.Run("AlreadyDeployed", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: false,
		}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		network := types.Network{}
		gateway := types.Gateway{
			Enabled: ptr.To(true),
		}

		status, err := cilium.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(status.Message).To(Equal(cilium.EnabledMsg))

		helmCiliumArgs := helmM.ApplyCalledWith[2]
		g.Expect(helmCiliumArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(helmCiliumArgs.State).To(Equal(helm.StateUpgradeOnly))
		g.Expect(helmCiliumArgs.Values["gatewayAPI"].(map[string]any)["enabled"]).To(Equal(true))

	})

	t.Run("RolloutFail", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset()
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
				KubernetesClient: &kubernetes.Client{
					Interface: clientset,
				},
			},
		}
		network := types.Network{}
		gateway := types.Gateway{
			Enabled: ptr.To(true),
		}

		status, err := cilium.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(status.Message).To(ContainSubstring("Failed to deploy Cilium Gateway"))
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
		gateway := types.Gateway{
			Enabled: ptr.To(true),
		}

		status, err := cilium.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(status.Message).To(Equal(cilium.EnabledMsg))
	})

}

func TestGatewayDisabled(t *testing.T) {
	t.Run("HelmApplyErr", func(t *testing.T) {
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
		network := types.Network{}
		gateway := types.Gateway{
			Enabled: ptr.To(false),
		}

		status, err := cilium.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.GatewayDeleteFailedMsgTmpl, applyErr)))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

	})

	t.Run("AlreadyDeleted", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: false,
		}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		network := types.Network{}
		gateway := types.Gateway{
			Enabled: ptr.To(false),
		}
		status, err := cilium.ApplyGateway(context.Background(), snapM, gateway, network, nil)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(status.Message).To(Equal(cilium.DisabledMsg))

		helmCiliumArgs := helmM.ApplyCalledWith[1]
		g.Expect(helmCiliumArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(helmCiliumArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(helmCiliumArgs.Values["gatewayAPI"].(map[string]any)["enabled"]).To(Equal(false))

	})

	t.Run("RolloutFail", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset()
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
				KubernetesClient: &kubernetes.Client{
					Interface: clientset,
				},
			},
		}
		network := types.Network{}
		gateway := types.Gateway{
			Enabled: ptr.To(false),
		}
		status, err := cilium.ApplyGateway(context.Background(), snapM, gateway, network, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(status.Message).To(ContainSubstring("Failed to deploy Cilium Gateway"))
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
		gateway := types.Gateway{
			Enabled: ptr.To(false),
		}

		status, err := cilium.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(status.Message).To(Equal(cilium.DisabledMsg))
	})
}
