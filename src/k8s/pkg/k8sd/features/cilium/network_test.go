package cilium_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gtypes "github.com/onsi/gomega/types"
	"k8s.io/utils/ptr"

	"github.com/canonical/k8s/pkg/client/helm"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"
)

// NOTE(hue): status.Message is not checked sometimes to avoid unnecessary complexity
type (
	given struct {
		config      types.Network
		expectState helm.State
		helmError   error
	}
	expect struct {
		// tests for error
		err []gtypes.GomegaMatcher
		// tests for status
		status []gtypes.GomegaMatcher
		//	tests for helm
		helm []gtypes.GomegaMatcher
	}
)

func TestNetwork(t *testing.T) {
	helmErr := errors.New("failed to apply")
	for _, tc := range []struct {
		name   string
		given  given
		expect expect
	}{
		{
			name: "NetworkDisabledHelmApplyFails",
			given: given{
				config:      types.Network{},
				expectState: 0,
				helmError:   helmErr,
			},
			expect: expect{
				err: []gtypes.GomegaMatcher{
					MatchError(helmErr),
				},
				status: []gtypes.GomegaMatcher{
					MatchAllFields(Fields{
						"Enabled":   BeFalse(),
						"Message":   ContainSubstring(helmErr.Error()),
						"Version":   Equal(cilium.CiliumAgentImageTag),
						"UpdatedAt": Ignore(),
					}),
				},
				helm: []gtypes.GomegaMatcher{
					MatchFields(IgnoreExtras, Fields{
						"ApplyCalledWith": HaveLen(1),
					}),
					MatchFields(IgnoreExtras, Fields{
						"ApplyCalledWith": MatchElementsWithIndex(IndexIdentity, IgnoreExtras, Elements{
							"0": MatchFields(IgnoreExtras, Fields{
								"Chart":  Equal(cilium.ChartCilium),
								"State":  Equal(helm.StateDeleted),
								"Values": BeNil(),
							}),
						}),
					}),
				},
			},
		},
		{
			name: "NetworkDisabledSuccess",
			given: given{
				config: types.Network{
					Enabled: ptr.To(false),
				},
				expectState: 0,
			},
			expect: expect{
				err: []gtypes.GomegaMatcher{
					Not(HaveOccurred()),
				},
				status: []gtypes.GomegaMatcher{
					MatchAllFields(Fields{
						"Enabled":   BeFalse(),
						"Message":   Equal(cilium.DisabledMsg),
						"Version":   Equal(cilium.CiliumAgentImageTag),
						"UpdatedAt": Ignore(),
					}),
				},
				helm: []gtypes.GomegaMatcher{
					MatchFields(IgnoreExtras, Fields{
						"ApplyCalledWith": HaveLen(1),
					}),
					MatchFields(IgnoreExtras, Fields{
						"ApplyCalledWith": MatchElementsWithIndex(IndexIdentity, IgnoreExtras, Elements{
							"0": MatchFields(IgnoreExtras, Fields{
								"Chart":  Equal(cilium.ChartCilium),
								"State":  Equal(helm.StateDeleted),
								"Values": BeNil(),
							}),
						}),
					}),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			helmM := &helmmock.Mock{
				ApplyErr: tc.given.helmError,
			}
			snapM := &snapmock.Snap{
				Mock: snapmock.Mock{
					HelmClient: helmM,
				},
			}

			status, err := cilium.ApplyNetwork(context.Background(), snapM, tc.given.config, nil)
			for _, matcher := range tc.expect.err {
				g.Expect(err).To(matcher)
			}
			for _, matcher := range tc.expect.status {
				g.Expect(status).To(matcher)
			}
			for _, matcher := range tc.expect.helm {
				g.Expect(*helmM).To(matcher)
			}
		})
	}
}

func TestNetworkEnabled(t *testing.T) {
	t.Run("InvalidCIDR", func(t *testing.T) {
		g := NewWithT(t)

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

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(0))
	})
	t.Run("Strict", func(t *testing.T) {
		g := NewWithT(t)

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

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Message).To(Equal(cilium.EnabledMsg))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateNetworkValues(g, callArgs.Values, cfg, snapM)
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
		cfg := types.Network{
			Enabled: ptr.To(true),
			PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
		}

		status, err := cilium.ApplyNetwork(context.Background(), snapM, cfg, nil)

		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateNetworkValues(g, callArgs.Values, cfg, snapM)
	})
}

func validateNetworkValues(g Gomega, values map[string]any, cfg types.Network, snap snap.Snap) {

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
