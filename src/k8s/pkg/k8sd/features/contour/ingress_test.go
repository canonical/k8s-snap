package contour_test

import (
	"context"
	"errors"
	"testing"
	"time"

	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features/contour"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestIngressDisabled(t *testing.T) {
	t.Run("HelmFailed", func(t *testing.T) {
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
		ingress := types.Ingress{
			Enabled: ptr.To(false),
		}

		status, err := contour.ApplyIngress(context.Background(), snapM, ingress, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(contour.ContourIngressContourImageTag))
		g.Expect(status.Message).To(ContainSubstring("failed to uninstall ingress"))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))
	})

	t.Run("Success", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		network := types.Network{}
		ingress := types.Ingress{
			Enabled: ptr.To(false),
		}

		status, err := contour.ApplyIngress(context.Background(), snapM, ingress, network, nil)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(contour.ContourIngressContourImageTag))
		g.Expect(status.Message).To(ContainSubstring("disabled"))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))
	})
}

func TestIngressEnabled(t *testing.T) {
	t.Run("HelmFailed", func(t *testing.T) {
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
		ingress := types.Ingress{
			Enabled: ptr.To(true),
		}

		status, err := contour.ApplyIngress(context.Background(), snapM, ingress, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(contour.ContourIngressContourImageTag))
		g.Expect(status.Message).To(ContainSubstring("Failed to deploy Contour Ingress"))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))
	})

	t.Run("Success", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset(
			&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ck-ingress-contour-contour",
					Namespace: "projectcontour",
				},
			})
		fakeDiscovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
		g.Expect(ok).To(BeTrue())
		fakeDiscovery.Resources = []*metav1.APIResourceList{
			{
				GroupVersion: "projectcontour.io/v1alpha1",
				APIResources: []metav1.APIResource{
					{Name: "contourconfigurations"},
					{Name: "contourdeployments"},
					{Name: "extensionservices"},
				},
			},
			{
				GroupVersion: "projectcontour.io/v1",
				APIResources: []metav1.APIResource{
					{Name: "tlscertificatedelegations"},
					{Name: "httpproxies"},
				},
			}}
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		status, err := contour.ApplyIngress(ctx, snapM, ingress, network, nil)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(contour.ContourIngressContourImageTag))
		g.Expect(status.Message).To(ContainSubstring("enabled"))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(3))
		validateIngressValues(g, helmM.ApplyCalledWith[1].Values, ingress)
	})

	t.Run("SuccessWithEnabledProxyProtocol", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset(
			&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ck-ingress-contour-contour",
					Namespace: "projectcontour",
				},
			})
		fakeDiscovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
		g.Expect(ok).To(BeTrue())
		fakeDiscovery.Resources = []*metav1.APIResourceList{
			{
				GroupVersion: "projectcontour.io/v1alpha1",
				APIResources: []metav1.APIResource{
					{Name: "contourconfigurations"},
					{Name: "contourdeployments"},
					{Name: "extensionservices"},
				},
			},
			{
				GroupVersion: "projectcontour.io/v1",
				APIResources: []metav1.APIResource{
					{Name: "tlscertificatedelegations"},
					{Name: "httpproxies"},
				},
			}}
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
			Enabled:             ptr.To(true),
			EnableProxyProtocol: ptr.To(true),
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		status, err := contour.ApplyIngress(ctx, snapM, ingress, network, nil)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(contour.ContourIngressContourImageTag))
		g.Expect(status.Message).To(ContainSubstring("enabled"))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(3))
		validateIngressValues(g, helmM.ApplyCalledWith[1].Values, ingress)
	})

	t.Run("SuccessWithDefaultTLSSecret", func(t *testing.T) {
		g := NewWithT(t)
		defaultTLSSecret := "secret"

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset(
			&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ck-ingress-contour-contour",
					Namespace: "projectcontour",
				},
			})
		fakeDiscovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
		g.Expect(ok).To(BeTrue())
		fakeDiscovery.Resources = []*metav1.APIResourceList{

			{
				GroupVersion: "projectcontour.io/v1alpha1",
				APIResources: []metav1.APIResource{
					{Name: "contourconfigurations"},
					{Name: "contourdeployments"},
					{Name: "extensionservices"},
				},
			},
			{
				GroupVersion: "projectcontour.io/v1",
				APIResources: []metav1.APIResource{
					{Name: "tlscertificatedelegations"},
					{Name: "httpproxies"},
				},
			}}
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
			Enabled:          ptr.To(true),
			DefaultTLSSecret: ptr.To(defaultTLSSecret),
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		status, err := contour.ApplyIngress(ctx, snapM, ingress, network, nil)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(contour.ContourIngressContourImageTag))
		g.Expect(status.Message).To(ContainSubstring("enabled"))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(3))
		validateIngressValues(g, helmM.ApplyCalledWith[1].Values, ingress)
		g.Expect(helmM.ApplyCalledWith[2].Values["defaultTLSSecret"]).To(Equal(defaultTLSSecret))
	})

	t.Run("NoCR", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset()
		fakeDiscovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
		g.Expect(ok).To(BeTrue())
		fakeDiscovery.Resources = []*metav1.APIResourceList{
			{
				GroupVersion: "projectcontour.io/v1alpha1",
				APIResources: []metav1.APIResource{
					{Name: "contourconfigurations"},
					{Name: "contourdeployments"},
					{Name: "extensionservices"},
				},
			},
			{
				GroupVersion: "projectcontour.io/metav1",
				APIResources: []metav1.APIResource{},
			}}
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		status, err := contour.ApplyIngress(ctx, snapM, ingress, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to wait for required contour common CRDs"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(contour.ContourIngressContourImageTag))
		g.Expect(status.Message).To(ContainSubstring("Failed to deploy Contour Ingress"))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))
	})

	t.Run("NoDeployment", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset(
			&v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy",
					Namespace: "projectcontour",
				},
			})
		fakeDiscovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
		g.Expect(ok).To(BeTrue())
		fakeDiscovery.Resources = []*metav1.APIResourceList{
			{
				GroupVersion: "projectcontour.io/v1alpha1",
				APIResources: []metav1.APIResource{
					{Name: "contourconfigurations"},
					{Name: "contourdeployments"},
					{Name: "extensionservices"},
				},
			},
			{
				GroupVersion: "projectcontour.io/v1",
				APIResources: []metav1.APIResource{
					{Name: "tlscertificatedelegations"},
					{Name: "httpproxies"},
				},
			}}
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		status, err := contour.ApplyIngress(ctx, snapM, ingress, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to rollout restart contour to apply ingress"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(contour.ContourIngressContourImageTag))
		g.Expect(status.Message).To(ContainSubstring("Failed to deploy Contour Ingress"))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(2))
	})
}

func validateIngressValues(g Gomega, values map[string]interface{}, ingress types.Ingress) {
	projectcontour := values["contour"].(map[string]any)["image"].(map[string]any)
	g.Expect(projectcontour["repository"]).To(Equal(contour.ContourIngressContourImageRepo))
	g.Expect(projectcontour["tag"]).To(Equal(contour.ContourIngressContourImageTag))

	envoy := values["envoy"].(map[string]any)["image"].(map[string]any)
	g.Expect(envoy["repository"]).To(Equal(contour.ContourIngressEnvoyImageRepo))
	g.Expect(envoy["tag"]).To(Equal(contour.ContourIngressEnvoyImageTag))

	if ingress.GetEnableProxyProtocol() {
		g.Expect(values["contour"].(map[string]any)["extraArgs"].([]string)[0]).To(Equal("--use-proxy-protocol"))
	}
}
