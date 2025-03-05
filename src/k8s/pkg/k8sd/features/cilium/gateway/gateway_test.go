package gateway_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/client/helm/loader"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	cilium_gateway "github.com/canonical/k8s/pkg/k8sd/features/cilium/gateway"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestGatewayEnabled(t *testing.T) {
	cilium_gateway.GetNetworkManifest = func(ctx context.Context, state state.State) (*types.FeatureManifest, error) {
		return &cilium_network.Manifest, nil
	}
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
		cfg := types.ClusterConfig{
			Network: types.Network{},
			Gateway: types.Gateway{
				Enabled: ptr.To(true),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))

		base := features.NewReconciler(cilium_gateway.Manifest, snapM, mc, nil, func() {})
		reconciler := cilium_gateway.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium_network.Manifest.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium_gateway.GatewayDeployFailedMsgTmpl, err)))
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
		cfg := types.ClusterConfig{
			Network: types.Network{},
			Gateway: types.Gateway{
				Enabled: ptr.To(true),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))

		base := features.NewReconciler(cilium_gateway.Manifest, snapM, mc, nil, func() {})
		reconciler := cilium_gateway.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(cilium_network.Manifest.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(status.Message).To(Equal(cilium.EnabledMsg))

		helmCiliumArgs := helmM.ApplyCalledWith[2]
		g.Expect(helmCiliumArgs.Chart).To(Equal(cilium_network.Manifest.GetChart(cilium_network.CiliumChartName)))
		g.Expect(helmCiliumArgs.State).To(Equal(helm.StateUpgradeOnly))
		g.Expect(helmCiliumArgs.Values["gatewayAPI"].(map[string]any)["enabled"]).To(BeTrue())
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
		cfg := types.ClusterConfig{
			Network: types.Network{},
			Gateway: types.Gateway{
				Enabled: ptr.To(true),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))

		base := features.NewReconciler(cilium_gateway.Manifest, snapM, mc, nil, func() {})
		reconciler := cilium_gateway.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium_network.Manifest.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium_gateway.GatewayDeployFailedMsgTmpl, err)))
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
		cfg := types.ClusterConfig{
			Network: types.Network{},
			Gateway: types.Gateway{
				Enabled: ptr.To(true),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))

		base := features.NewReconciler(cilium_gateway.Manifest, snapM, mc, nil, func() {})
		reconciler := cilium_gateway.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(cilium_network.Manifest.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(status.Message).To(Equal(cilium.EnabledMsg))
	})
}

func TestGatewayDisabled(t *testing.T) {
	cilium_gateway.GetNetworkManifest = func(ctx context.Context, state state.State) (*types.FeatureManifest, error) {
		return &cilium_network.Manifest, nil
	}
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
		cfg := types.ClusterConfig{
			Network: types.Network{},
			Gateway: types.Gateway{
				Enabled: ptr.To(false),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))

		base := features.NewReconciler(cilium_gateway.Manifest, snapM, mc, nil, func() {})
		reconciler := cilium_gateway.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium_network.Manifest.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium_gateway.GatewayDeleteFailedMsgTmpl, err)))
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
		cfg := types.ClusterConfig{
			Network: types.Network{},
			Gateway: types.Gateway{
				Enabled: ptr.To(false),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))

		base := features.NewReconciler(cilium_gateway.Manifest, snapM, mc, nil, func() {})
		reconciler := cilium_gateway.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium_network.Manifest.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(status.Message).To(Equal(cilium.DisabledMsg))

		helmCiliumArgs := helmM.ApplyCalledWith[1]
		g.Expect(helmCiliumArgs.Chart).To(Equal(cilium_network.Manifest.GetChart(cilium_network.CiliumChartName)))
		g.Expect(helmCiliumArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(helmCiliumArgs.Values["gatewayAPI"].(map[string]any)["enabled"]).To(BeFalse())
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
		cfg := types.ClusterConfig{
			Network: types.Network{},
			Gateway: types.Gateway{
				Enabled: ptr.To(false),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))

		base := features.NewReconciler(cilium_gateway.Manifest, snapM, mc, nil, func() {})
		reconciler := cilium_gateway.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium_network.Manifest.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium_gateway.GatewayDeployFailedMsgTmpl, err)))
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
		cfg := types.ClusterConfig{
			Network: types.Network{},
			Gateway: types.Gateway{
				Enabled: ptr.To(false),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&cilium.ChartFS))

		base := features.NewReconciler(cilium_gateway.Manifest, snapM, mc, nil, func() {})
		reconciler := cilium_gateway.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium_network.Manifest.GetImage(cilium_network.CiliumAgentImageName).Tag))
		g.Expect(status.Message).To(Equal(cilium.DisabledMsg))
	})
}
