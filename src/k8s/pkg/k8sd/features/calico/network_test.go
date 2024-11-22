package calico_test

import (
	"context"
	"errors"
	"testing"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/calico"
	"github.com/canonical/k8s/pkg/client/helm"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/k8sd/features/calico"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

// NOTE(hue): status.Message is not checked sometimes to avoid unnecessary complexity

var defaultAnnotations = types.Annotations{
	apiv1_annotations.AnnotationAPIServerEnabled:          "true",
	apiv1_annotations.AnnotationEncapsulationV4:           "VXLAN",
	apiv1_annotations.AnnotationEncapsulationV6:           "VXLAN",
	apiv1_annotations.AnnotationAutodetectionV4FirstFound: "true",
	apiv1_annotations.AnnotationAutodetectionV6FirstFound: "true",
}

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
		network := types.Network{
			Enabled: ptr.To(false),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := calico.ApplyNetwork(context.Background(), snapM, "127.0.0.1", apiserver, network, nil)

		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Version).To(Equal(calico.CalicoTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(calico.ChartCalico))
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
		network := types.Network{
			Enabled: ptr.To(false),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := calico.ApplyNetwork(context.Background(), snapM, "127.0.0.1", apiserver, network, nil)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(calico.DisabledMsg))
		g.Expect(status.Version).To(Equal(calico.CalicoTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(calico.ChartCalico))
		g.Expect(callArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(callArgs.Values).To(BeNil())
	})
}

func TestEnabled(t *testing.T) {
	t.Run("InvalidPodCIDR", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		network := types.Network{
			Enabled: ptr.To(true),
			PodCIDR: ptr.To("invalid-cidr"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := calico.ApplyNetwork(context.Background(), snapM, "127.0.0.1", apiserver, network, defaultAnnotations)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(calico.CalicoTag))
		g.Expect(helmM.ApplyCalledWith).To(BeEmpty())
	})
	t.Run("InvalidServiceCIDR", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		network := types.Network{
			Enabled:     ptr.To(true),
			PodCIDR:     ptr.To("192.0.2.0/24,2001:db8::/32"),
			ServiceCIDR: ptr.To("invalid-cidr"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := calico.ApplyNetwork(context.Background(), snapM, "127.0.0.1", apiserver, network, defaultAnnotations)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(calico.CalicoTag))
		g.Expect(helmM.ApplyCalledWith).To(BeEmpty())
	})
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
		network := types.Network{
			Enabled:     ptr.To(true),
			PodCIDR:     ptr.To("192.0.2.0/24,2001:db8::/32"),
			ServiceCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := calico.ApplyNetwork(context.Background(), snapM, "127.0.0.1", apiserver, network, defaultAnnotations)

		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Version).To(Equal(calico.CalicoTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(calico.ChartCalico))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(t, callArgs.Values, network)
	})
	t.Run("Success", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		network := types.Network{
			Enabled:     ptr.To(true),
			PodCIDR:     ptr.To("192.0.2.0/24,2001:db8::/32"),
			ServiceCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := calico.ApplyNetwork(context.Background(), snapM, "127.0.0.1", apiserver, network, defaultAnnotations)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Message).To(Equal(calico.EnabledMsg))
		g.Expect(status.Version).To(Equal(calico.CalicoTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(calico.ChartCalico))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(t, callArgs.Values, network)
	})
}

func validateValues(t *testing.T, values map[string]any, network types.Network) {
	g := NewWithT(t)

	podIPv4CIDR, podIPv6CIDR, err := utils.SplitCIDRStrings(network.GetPodCIDR())
	g.Expect(err).ToNot(HaveOccurred())

	svcIPv4CIDR, svcIPv6CIDR, err := utils.SplitCIDRStrings(network.GetServiceCIDR())
	g.Expect(err).ToNot(HaveOccurred())

	// calico network
	calicoNetwork := values["installation"].(map[string]any)["calicoNetwork"].(map[string]any)
	g.Expect(calicoNetwork["ipPools"].([]map[string]any)).To(ContainElements(map[string]any{
		"name":          "ipv4-ippool",
		"cidr":          podIPv4CIDR,
		"encapsulation": "VXLAN",
	}))
	g.Expect(calicoNetwork["ipPools"].([]map[string]any)).To(ContainElements(map[string]any{
		"name":          "ipv6-ippool",
		"cidr":          podIPv6CIDR,
		"encapsulation": "VXLAN",
	}))
	g.Expect(calicoNetwork["ipPools"].([]map[string]any)).To(HaveLen(2))
	g.Expect(calicoNetwork["nodeAddressAutodetectionV4"].(map[string]any)["firstFound"]).To(BeTrue())
	g.Expect(calicoNetwork["nodeAddressAutodetectionV6"].(map[string]any)["firstFound"]).To(BeTrue())

	g.Expect(values["apiServer"].(map[string]any)["enabled"]).To(BeTrue())

	// service CIDRs
	g.Expect(values["serviceCIDRs"].([]string)).To(ContainElements(svcIPv4CIDR, svcIPv6CIDR))
	g.Expect(values["serviceCIDRs"].([]string)).To(HaveLen(2))
}
