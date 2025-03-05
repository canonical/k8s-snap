package dns_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/client/helm/loader"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	coredns_dns "github.com/canonical/k8s/pkg/k8sd/features/coredns/dns"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestDisabled(t *testing.T) {
	coredns_dns.UpdateClusterDNS = func(ctx context.Context, s state.State, dnsIP string) error {
		return nil
	}
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
		cfg := types.ClusterConfig{
			DNS: types.DNS{
				Enabled: ptr.To(false),
			},
			Kubelet: types.Kubelet{},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		base := features.NewReconciler(coredns_dns.Manifest, snapM, mc, nil, func() {})
		reconciler := coredns_dns.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(MatchError(ContainSubstring(applyErr.Error())))
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Message).To(ContainSubstring("failed to uninstall coredns"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns_dns.Manifest.GetImage(coredns_dns.CoreDNSImageName).Tag))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.Manifest.GetChart(coredns_dns.CoreDNSChartName)))
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
		cfg := types.ClusterConfig{
			DNS: types.DNS{
				Enabled: ptr.To(false),
			},
			Kubelet: types.Kubelet{},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		base := features.NewReconciler(coredns_dns.Manifest, snapM, mc, nil, func() {})
		reconciler := coredns_dns.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(status.Message).To(Equal("disabled"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns_dns.Manifest.GetImage(coredns_dns.CoreDNSImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.Manifest.GetChart(coredns_dns.CoreDNSChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(callArgs.Values).To(BeNil())
	})
}

func TestEnabled(t *testing.T) {
	coredns_dns.UpdateClusterDNS = func(ctx context.Context, s state.State, dnsIP string) error {
		return nil
	}
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
		cfg := types.ClusterConfig{
			DNS: types.DNS{
				Enabled: ptr.To(true),
			},
			Kubelet: types.Kubelet{},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		base := features.NewReconciler(coredns_dns.Manifest, snapM, mc, nil, func() {})
		reconciler := coredns_dns.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(MatchError(ContainSubstring(applyErr.Error())))
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Message).To(ContainSubstring("failed to apply coredns"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns_dns.Manifest.GetImage(coredns_dns.CoreDNSImageName).Tag))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.Manifest.GetChart(coredns_dns.CoreDNSChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(g, callArgs.Values, cfg.DNS, cfg.Kubelet)
	})
	t.Run("HelmApplySuccessServiceFails", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{}
		clientset := fake.NewSimpleClientset()
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient:       helmM,
				KubernetesClient: &kubernetes.Client{Interface: clientset},
			},
		}
		cfg := types.ClusterConfig{
			DNS: types.DNS{
				Enabled: ptr.To(true),
			},
			Kubelet: types.Kubelet{},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		base := features.NewReconciler(coredns_dns.Manifest, snapM, mc, nil, func() {})
		reconciler := coredns_dns.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(MatchError(ContainSubstring("services \"coredns\" not found")))
		g.Expect(status.Message).To(ContainSubstring("failed to retrieve the coredns service"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns_dns.Manifest.GetImage(coredns_dns.CoreDNSImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.Manifest.GetChart(coredns_dns.CoreDNSChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(g, callArgs.Values, cfg.DNS, cfg.Kubelet)
	})
	t.Run("Success", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{}
		clusterIp := "10.96.0.10"
		corednsService := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "coredns",
				Namespace: "kube-system",
			},
			Spec: corev1.ServiceSpec{
				ClusterIP: clusterIp,
			},
		}
		clientset := fake.NewSimpleClientset(corednsService)
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient:       helmM,
				KubernetesClient: &kubernetes.Client{Interface: clientset},
			},
		}
		cfg := types.ClusterConfig{
			Network: types.Network{
				PodCIDR: ptr.To("10.96.0.0/24"),
			},
			DNS: types.DNS{
				Enabled: ptr.To(true),
			},
			Kubelet: types.Kubelet{},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		base := features.NewReconciler(coredns_dns.Manifest, snapM, mc, nil, func() {})
		reconciler := coredns_dns.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(status.Message).To(ContainSubstring("enabled at " + clusterIp))
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(coredns_dns.Manifest.GetImage(coredns_dns.CoreDNSImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.Manifest.GetChart(coredns_dns.CoreDNSChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(g, callArgs.Values, cfg.DNS, cfg.Kubelet)
	})
}

func validateValues(g Gomega, values map[string]any, dns types.DNS, kubelet types.Kubelet) {
	service := values["service"].(map[string]any)
	g.Expect(service["clusterIP"]).To(Equal(kubelet.GetClusterDNS()))

	servers := values["servers"].([]map[string]any)
	plugins := servers[0]["plugins"].([]map[string]any)
	g.Expect(plugins[3]["parameters"]).To(ContainSubstring(kubelet.GetClusterDomain()))
	g.Expect(plugins[5]["parameters"]).To(ContainSubstring(strings.Join(dns.GetUpstreamNameservers(), " ")))
}
