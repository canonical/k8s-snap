package cilium_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/canonical/k8s/pkg/client/helm"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"
	"k8s.io/utils/ptr"
)

// NOTE(hue): status.Message is not checked sometimes to avoid unnecessary complexity

func TestDisabled(t *testing.T) {
	t.Run("HelmApplyFails", func(t *testing.T) {
		applyErr := errors.New("failed to apply")
		helmM := &helmmock.Mock{
			ApplyErr: applyErr,
		}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		cfg := types.Network{
			Enabled: ptr.To(false),
		}

		status, err := cilium.ApplyNetwork(context.Background(), snapM, cfg, nil)

		g := NewWithT(t)
		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(callArgs.Values).To(BeNil())
	})
	t.Run("Success", func(t *testing.T) {
		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		cfg := types.Network{
			Enabled: ptr.To(false),
		}

		status, err := cilium.ApplyNetwork(context.Background(), snapM, cfg, nil)

		g := NewWithT(t)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(cilium.DisabledMsg))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(callArgs.Values).To(BeNil())
	})
}

func TestEnabled(t *testing.T) {
	t.Run("InvalidCIDR", func(t *testing.T) {
		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		cfg := types.Network{
			Enabled: ptr.To(true),
			PodCIDR: ptr.To("invalid-cidr"),
		}

		status, err := cilium.ApplyNetwork(context.Background(), snapM, cfg, nil)

		g := NewWithT(t)
		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(0))
	})
	t.Run("Strict", func(t *testing.T) {
		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
				Strict:     true,
			},
		}
		cfg := types.Network{
			Enabled: ptr.To(true),
			PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
		}

		status, err := cilium.ApplyNetwork(context.Background(), snapM, cfg, nil)

		g := NewWithT(t)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Message).To(Equal(cilium.EnabledMsg))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(t, callArgs.Values, cfg, snapM)
	})
	t.Run("HelmApplyFails", func(t *testing.T) {
		applyErr := errors.New("failed to apply")
		helmM := &helmmock.Mock{
			ApplyErr: applyErr,
		}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		cfg := types.Network{
			Enabled: ptr.To(true),
			PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
		}

		status, err := cilium.ApplyNetwork(context.Background(), snapM, cfg, nil)

		g := NewWithT(t)
		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateValues(t, callArgs.Values, cfg, snapM)
	})
}

func validateValues(t *testing.T, values map[string]any, cfg types.Network, snap snap.Snap) {
	t.Helper()
	g := NewWithT(t)

	ipv4CIDR, ipv6CIDR, err := utils.ParseCIDRs(cfg.GetPodCIDR())
	g.Expect(err).ToNot(HaveOccurred())

	bpfMount, err := utils.GetMountPath("bpf")
	g.Expect(err).ToNot(HaveOccurred())

	cgrMount, err := utils.GetMountPath("cgroup2")
	g.Expect(err).ToNot(HaveOccurred())

	if snap.Strict() {
		g.Expect(values["bpf"].(map[string]any)["root"]).To(Equal(bpfMount))
		g.Expect(values["cgroup"].(map[string]any)["hostRoot"]).To(Equal(cgrMount))
	}

	g.Expect(values["ipam"].(map[string]any)["operator"].(map[string]any)["clusterPoolIPv4PodCIDRList"]).To(Equal(ipv4CIDR))
	g.Expect(values["ipam"].(map[string]any)["operator"].(map[string]any)["clusterPoolIPv6PodCIDRList"]).To(Equal(ipv6CIDR))
	g.Expect(values["ipv4"].(map[string]any)["enabled"]).To(Equal((ipv4CIDR != "")))
	g.Expect(values["ipv6"].(map[string]any)["enabled"]).To(Equal((ipv6CIDR != "")))
}
