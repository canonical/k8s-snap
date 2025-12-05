package coredns_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/canonical/k8s/pkg/client/helm"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
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

		status, str, err := coredns.ApplyDNS(context.Background(), snapM, dns, kubelet, nil)

		g.Expect(err).To(MatchError(ContainSubstring(applyErr.Error())))
		g.Expect(str).To(BeEmpty())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Message).To(ContainSubstring("failed to uninstall coredns"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns.ImageTag))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns.Chart))
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

		status, str, err := coredns.ApplyDNS(context.Background(), snapM, dns, kubelet, nil)

		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(str).To(BeEmpty())
		g.Expect(status.Message).To(Equal("disabled"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns.ImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns.Chart))
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

		status, str, err := coredns.ApplyDNS(context.Background(), snapM, dns, kubelet, nil)

		g.Expect(err).To(MatchError(ContainSubstring(applyErr.Error())))
		g.Expect(str).To(BeEmpty())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Message).To(ContainSubstring("failed to apply coredns"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns.ImageTag))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns.Chart))
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

		status, str, err := coredns.ApplyDNS(context.Background(), snapM, dns, kubelet, nil)

		g.Expect(err).To(MatchError(ContainSubstring("services \"coredns\" not found")))
		g.Expect(str).To(BeEmpty())
		g.Expect(status.Message).To(ContainSubstring("failed to retrieve the coredns service"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(coredns.ImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns.Chart))
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

		status, str, err := coredns.ApplyDNS(context.Background(), snapM, dns, kubelet, nil)

		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(str).To(Equal(clusterIp))
		g.Expect(status.Message).To(ContainSubstring("enabled at " + clusterIp))
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(coredns.ImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(coredns.Chart))
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

	// Validate PriorityClass
	g.Expect(values["priorityClassName"]).To(Equal("system-node-critical"))

	// Validate tolerations
	tolerations, ok := values["tolerations"].([]map[string]any)
	g.Expect(ok).To(BeTrue())
	g.Expect(tolerations).To(HaveLen(1))

	toleration := tolerations[0]
	g.Expect(toleration["key"]).To(Equal("node-role.kubernetes.io/control-plane"))
	g.Expect(toleration["operator"]).To(Equal("Exists"))
	g.Expect(toleration["effect"]).To(Equal("NoSchedule"))

	// Validate HPA configuration
	hpa := values["hpa"].(map[string]any)
	g.Expect(hpa["enabled"]).To(BeTrue())
	g.Expect(hpa["minReplicas"]).To(Equal(2))
	g.Expect(hpa["maxReplicas"]).To(Equal(100))

	metrics := hpa["metrics"].([]map[string]any)
	g.Expect(metrics).To(HaveLen(2))

	// CPU metric
	g.Expect(metrics[0]["type"]).To(Equal("Resource"))
	cpuResource := metrics[0]["resource"].(map[string]any)
	g.Expect(cpuResource["name"]).To(Equal("cpu"))
	cpuTarget := cpuResource["target"].(map[string]any)
	g.Expect(cpuTarget["type"]).To(Equal("Utilization"))
	g.Expect(cpuTarget["averageUtilization"]).To(Equal(80))

	// Memory metric
	g.Expect(metrics[1]["type"]).To(Equal("Resource"))
	memResource := metrics[1]["resource"].(map[string]any)
	g.Expect(memResource["name"]).To(Equal("memory"))
	memTarget := memResource["target"].(map[string]any)
	g.Expect(memTarget["type"]).To(Equal("Utilization"))
	g.Expect(memTarget["averageUtilization"]).To(Equal(70))

	// Validate PDB configuration
	pdb := values["podDisruptionBudget"].(map[string]any)
	g.Expect(pdb["minAvailable"]).To(Equal(1))

	// Validate PodAntiAffinity
	affinity := values["affinity"].(map[string]any)
	podAntiAffinity := affinity["podAntiAffinity"].(map[string]any)
	preferred := podAntiAffinity["preferredDuringSchedulingIgnoredDuringExecution"].([]map[string]any)
	g.Expect(preferred).To(HaveLen(1))
	g.Expect(preferred[0]["weight"]).To(Equal(100))

	podAffinityTerm := preferred[0]["podAffinityTerm"].(map[string]any)
	g.Expect(podAffinityTerm["topologyKey"]).To(Equal("kubernetes.io/hostname"))
	labelSelector := podAffinityTerm["labelSelector"].(map[string]any)
	matchLabels := labelSelector["matchLabels"].(map[string]any)
	g.Expect(matchLabels["app.kubernetes.io/name"]).To(Equal("coredns"))
	g.Expect(matchLabels["app.kubernetes.io/instance"]).To(Equal("ck-dns"))

	// Validate TopologySpreadConstraints
	topologySpread := values["topologySpreadConstraints"].([]map[string]any)
	g.Expect(topologySpread).To(HaveLen(2))
	// Zone constraint
	zoneSelector := topologySpread[0]["labelSelector"].(map[string]any)
	zoneMatchLabels := zoneSelector["matchLabels"].(map[string]any)
	g.Expect(zoneMatchLabels["app.kubernetes.io/name"]).To(Equal("coredns"))
	g.Expect(zoneMatchLabels["app.kubernetes.io/instance"]).To(Equal("ck-dns"))
	g.Expect(zoneMatchLabels["k8s-app"]).To(Equal("coredns"))

	// Hostname constraint
	hostnameSelector := topologySpread[1]["labelSelector"].(map[string]any)
	hostnameMatchLabels := hostnameSelector["matchLabels"].(map[string]any)
	g.Expect(hostnameMatchLabels["app.kubernetes.io/name"]).To(Equal("coredns"))
	g.Expect(hostnameMatchLabels["app.kubernetes.io/instance"]).To(Equal("ck-dns"))
	g.Expect(hostnameMatchLabels["k8s-app"]).To(Equal("coredns"))

	// Zone constraint
	g.Expect(topologySpread[0]["maxSkew"]).To(Equal(1))
	g.Expect(topologySpread[0]["topologyKey"]).To(Equal("topology.kubernetes.io/zone"))
	g.Expect(topologySpread[0]["whenUnsatisfiable"]).To(Equal("ScheduleAnyway"))
	zoneMatchLabelKeys, ok := topologySpread[0]["matchLabelKeys"].([]string)
	g.Expect(ok).To(BeTrue())
	g.Expect(zoneMatchLabelKeys).To(Equal([]string{"pod-template-hash"}))

	// Hostname constraint
	g.Expect(topologySpread[1]["maxSkew"]).To(Equal(1))
	g.Expect(topologySpread[1]["topologyKey"]).To(Equal("kubernetes.io/hostname"))
	g.Expect(topologySpread[1]["whenUnsatisfiable"]).To(Equal("DoNotSchedule"))
	hostnameMatchLabelKeys, ok := topologySpread[1]["matchLabelKeys"].([]string)
	g.Expect(ok).To(BeTrue())
	g.Expect(hostnameMatchLabelKeys).To(Equal([]string{"pod-template-hash"}))
}
