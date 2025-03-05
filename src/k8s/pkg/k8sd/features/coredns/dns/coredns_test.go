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
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	coredns_dns "github.com/canonical/k8s/pkg/k8sd/features/coredns/dns"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestDisabled(t *testing.T) {
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
		dns := types.DNS{
			Enabled: ptr.To(false),
		}
		kubelet := types.Kubelet{}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		reconciler := coredns_dns.NewDNSReconciler(snapM, mc, nil)

		status, str, err := reconciler.ApplyDNS(context.Background(), dns, kubelet, nil)

		g.Expect(err).To(MatchError(ContainSubstring(applyErr.Error())))
		g.Expect(str).To(BeEmpty())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Message).To(ContainSubstring("failed to uninstall coredns"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns_dns.FeatureDNS.GetImage(coredns_dns.CoreDNSImageName).Tag))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.FeatureDNS.GetChart(coredns_dns.CoreDNSChartName)))
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
		dns := types.DNS{
			Enabled: ptr.To(false),
		}
		kubelet := types.Kubelet{}
		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		reconciler := coredns_dns.NewDNSReconciler(snapM, mc, nil)

		status, str, err := reconciler.ApplyDNS(context.Background(), dns, kubelet, nil)

		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(str).To(BeEmpty())
		g.Expect(status.Message).To(Equal("disabled"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns_dns.FeatureDNS.GetImage(coredns_dns.CoreDNSImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.FeatureDNS.GetChart(coredns_dns.CoreDNSChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(callArgs.Values).To(BeNil())
	})
}

func TestEnabled(t *testing.T) {
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
		dns := types.DNS{
			Enabled: ptr.To(true),
		}
		kubelet := types.Kubelet{}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		reconciler := coredns_dns.NewDNSReconciler(snapM, mc, nil)

		status, str, err := reconciler.ApplyDNS(context.Background(), dns, kubelet, nil)

		g.Expect(err).To(MatchError(ContainSubstring(applyErr.Error())))
		g.Expect(str).To(BeEmpty())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Message).To(ContainSubstring("failed to apply coredns"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns_dns.FeatureDNS.GetImage(coredns_dns.CoreDNSImageName).Tag))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.FeatureDNS.GetChart(coredns_dns.CoreDNSChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(g, callArgs.Values, dns, kubelet)
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
		dns := types.DNS{
			Enabled: ptr.To(true),
		}
		kubelet := types.Kubelet{}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		reconciler := coredns_dns.NewDNSReconciler(snapM, mc, nil)

		status, str, err := reconciler.ApplyDNS(context.Background(), dns, kubelet, nil)

		g.Expect(err).To(MatchError(ContainSubstring("services \"coredns\" not found")))
		g.Expect(str).To(BeEmpty())
		g.Expect(status.Message).To(ContainSubstring("failed to retrieve the coredns service"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns_dns.FeatureDNS.GetImage(coredns_dns.CoreDNSImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.FeatureDNS.GetChart(coredns_dns.CoreDNSChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(g, callArgs.Values, dns, kubelet)
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
		dns := types.DNS{
			Enabled: ptr.To(true),
		}
		kubelet := types.Kubelet{}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&coredns.ChartFS))

		reconciler := coredns_dns.NewDNSReconciler(snapM, mc, nil)

		status, str, err := reconciler.ApplyDNS(context.Background(), dns, kubelet, nil)

		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(str).To(Equal(clusterIp))
		g.Expect(status.Message).To(ContainSubstring("enabled at " + clusterIp))
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(coredns_dns.FeatureDNS.GetImage(coredns_dns.CoreDNSImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns_dns.FeatureDNS.GetChart(coredns_dns.CoreDNSChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(g, callArgs.Values, dns, kubelet)
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
