package contour_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features/contour"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestGatewayDisabled(t *testing.T) {
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
		gateway := types.Gateway{
			Enabled: ptr.To(false),
		}

		status, err := contour.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(contour.ContourGatewayProvisionerContourImageTag))
		g.Expect(status.Message).To(Equal(fmt.Sprintf(contour.GatewayDeleteFailedMsgTmpl,
			fmt.Errorf("failed to uninstall the contour gateway chart: %w", applyErr),
		)))
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
		gateway := types.Gateway{
			Enabled: ptr.To(false),
		}

		status, err := contour.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(contour.ContourGatewayProvisionerContourImageTag))
		g.Expect(status.Message).To(Equal(contour.DisabledMsg))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

	})
}

func TestGatewayEnabled(t *testing.T) {
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
		gateway := types.Gateway{
			Enabled: ptr.To(true),
		}

		status, err := contour.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(contour.ContourGatewayProvisionerContourImageTag))
		g.Expect(status.Message).To(Equal(fmt.Sprintf(contour.GatewayDeployFailedMsgTmpl,
			fmt.Errorf("failed to apply common contour CRDS: %w", fmt.Errorf("failed to install common CRDS: %w", applyErr)),
		)))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))
	})

	t.Run("Success", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset()
		fakeDiscovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
		g.Expect(ok).To(BeTrue())
		fakeDiscovery.Resources = []*v1.APIResourceList{
			{
				GroupVersion: "projectcontour.io/v1alpha1",
				APIResources: []v1.APIResource{
					{Name: "contourconfigurations"},
					{Name: "contourdeployments"},
					{Name: "extensionservices"},
				},
			},
			{
				GroupVersion: "projectcontour.io/v1",
				APIResources: []v1.APIResource{
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
		gateway := types.Gateway{
			Enabled: ptr.To(true),
		}

		status, err := contour.ApplyGateway(context.Background(), snapM, gateway, network, nil)

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(contour.ContourGatewayProvisionerContourImageTag))
		g.Expect(status.Message).To(Equal(contour.EnabledMsg))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(2))

		values := helmM.ApplyCalledWith[1].Values
		projectcontour := values["projectcontour"].(map[string]any)["image"].(map[string]any)
		g.Expect(projectcontour["repository"]).To(Equal(contour.ContourGatewayProvisionerContourImageRepo))
		g.Expect(projectcontour["tag"]).To(Equal(contour.ContourGatewayProvisionerContourImageTag))
		envoy := values["envoyproxy"].(map[string]any)["image"].(map[string]any)
		g.Expect(envoy["repository"]).To(Equal(contour.ContourGatewayProvisionerEnvoyImageRepo))
		g.Expect(envoy["tag"]).To(Equal(contour.ContourGatewayProvisionerEnvoyImageTag))
	})

	t.Run("CrdDeploymentFailed", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{
			ApplyChanged: true,
		}
		clientset := fake.NewSimpleClientset()
		fakeDiscovery, ok := clientset.Discovery().(*fakediscovery.FakeDiscovery)
		g.Expect(ok).To(BeTrue())
		fakeDiscovery.Resources = []*v1.APIResourceList{
			{
				GroupVersion: "projectcontour.io/v1alpha1",
				APIResources: []v1.APIResource{
					{Name: "contourconfigurations"},
					{Name: "contourdeployments"},
					{Name: "extensionservices"},
				},
			},
			{
				GroupVersion: "projectcontour.io/v1",
				APIResources: []v1.APIResource{},
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
		gateway := types.Gateway{
			Enabled: ptr.To(true),
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		status, err := contour.ApplyGateway(ctx, snapM, gateway, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to wait for required contour common CRDs"))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(contour.ContourGatewayProvisionerContourImageTag))
		g.Expect(status.Message).To(Equal(fmt.Sprintf(contour.GatewayDeployFailedMsgTmpl,
			fmt.Errorf("failed to wait for required contour common CRDs to be available: %w",
				errors.New("context deadline exceeded")),
		)))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))
	})
}
