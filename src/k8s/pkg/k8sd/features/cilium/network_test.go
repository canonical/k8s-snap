package cilium

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/client/helm"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"

	. "github.com/onsi/gomega"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
	"k8s.io/utils/ptr"
)

// NOTE(hue): status.Message is not checked sometimes to avoid unnecessary complexity
func TestNetworkDisabled(t *testing.T) {
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

		status, err := ApplyNetwork(context.Background(), snapM, apiserver, network, nil)

		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(networkDeleteFailedMsgTmpl, err)))
		g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(ChartCilium))
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

		status, err := ApplyNetwork(context.Background(), snapM, apiserver, network, nil)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(DisabledMsg))
		g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StateDeleted))
		g.Expect(callArgs.Values).To(BeNil())
	})
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
		network := types.Network{
			Enabled: ptr.To(true),
			PodCIDR: ptr.To("invalid-cidr"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := ApplyNetwork(context.Background(), snapM, apiserver, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
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
		network := types.Network{
			Enabled:          ptr.To(true),
			LocalhostAddress: ptr.To("127.0.0.1"),
			PodCIDR:          ptr.To("192.0.2.0/24,2001:db8::/32"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := ApplyNetwork(context.Background(), snapM, apiserver, network, nil)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Message).To(Equal(EnabledMsg))
		g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateNetworkValues(g, callArgs.Values, network, snapM)
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
			Enabled:          ptr.To(true),
			LocalhostAddress: ptr.To("127.0.0.1"),
			PodCIDR:          ptr.To("192.0.2.0/24,2001:db8::/32"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := ApplyNetwork(context.Background(), snapM, apiserver, network, nil)

		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(networkDeployFailedMsgTmpl, err)))
		g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(ChartCilium))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))
		validateNetworkValues(g, callArgs.Values, network, snapM)
	})
}

func TestNetworkMountPath(t *testing.T) {
	for _, tc := range []struct {
		name string
	}{
		{name: "bpf"},
		{name: "cgroup2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			mountPathErr := fmt.Errorf("%s not found", tc.name)
			helmM := &helmmock.Mock{}
			snapM := &snapmock.Snap{
				Mock: snapmock.Mock{
					HelmClient: helmM,
					Strict:     true,
				},
			}
			network := types.Network{
				Enabled:          ptr.To(true),
				LocalhostAddress: ptr.To("127.0.0.1"),
				PodCIDR:          ptr.To("192.0.2.0/24,2001:db8::/32"),
			}
			apiserver := types.APIServer{
				SecurePort: ptr.To(6443),
			}
			getMountPath = func(fsType string) (string, error) {
				if fsType == tc.name {
					return "", mountPathErr
				}
				return tc.name, nil
			}

			status, err := ApplyNetwork(context.Background(), snapM, apiserver, network, nil)

			g.Expect(err).To(HaveOccurred())
			g.Expect(err).To(MatchError(mountPathErr))
			g.Expect(status.Enabled).To(BeFalse())
			g.Expect(status.Message).To(Equal(fmt.Sprintf(networkDeployFailedMsgTmpl, err)))
			g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(HaveLen(0))

		})
	}
}

func TestNetworkMountPropagationType(t *testing.T) {
	t.Run("failedGetMountSys", func(t *testing.T) {
		g := NewWithT(t)

		mountErr := errors.New("/sys not found")
		getMountPropagationType = func(path string) (utils.MountPropagationType, error) {
			return "", mountErr
		}
		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
				Strict:     false,
			},
		}
		network := types.Network{
			Enabled:          ptr.To(true),
			LocalhostAddress: ptr.To("127.0.0.1"),
			PodCIDR:          ptr.To("192.0.2.0/24,2001:db8::/32"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := ApplyNetwork(context.Background(), snapM, apiserver, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(mountErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(networkDeployFailedMsgTmpl, err)))

		g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(0))

	})

	t.Run("MountPropagationPrivateOnLXDError", func(t *testing.T) {
		g := NewWithT(t)

		getMountPropagationType = func(path string) (utils.MountPropagationType, error) {
			return utils.MountPropagationPrivate, nil
		}
		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
				Strict:     false,
				OnLXDErr:   errors.New("failed to check LXD"),
			},
		}
		network := types.Network{
			Enabled:          ptr.To(true),
			LocalhostAddress: ptr.To("127.0.0.1"),
			PodCIDR:          ptr.To("192.0.2.0/24,2001:db8::/32"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}
		logger := ktesting.NewLogger(t, ktesting.NewConfig(ktesting.BufferLogs(true)))
		ctx := klog.NewContext(context.Background(), logger)

		status, err := ApplyNetwork(ctx, snapM, apiserver, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(networkDeployFailedMsgTmpl, err)))

		g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(0))
		testingLogger, ok := logger.GetSink().(ktesting.Underlier)
		if !ok {
			panic("Should have had a ktesting LogSink!?")
		}
		g.Expect(testingLogger.GetBuffer().String()).To(ContainSubstring("Failed to check if running on LXD"))
	})

	t.Run("MountPropagationPrivateOnLXD", func(t *testing.T) {
		g := NewWithT(t)

		getMountPropagationType = func(path string) (utils.MountPropagationType, error) {
			return utils.MountPropagationPrivate, nil
		}
		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
				Strict:     false,
				OnLXD:      true,
			},
		}
		network := types.Network{
			Enabled:          ptr.To(true),
			LocalhostAddress: ptr.To("127.0.0.1"),
			PodCIDR:          ptr.To("192.0.2.0/24,2001:db8::/32"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := ApplyNetwork(context.Background(), snapM, apiserver, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(networkDeployFailedMsgTmpl, err)))

		g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(0))
	})

	t.Run("MountPropagationPrivate", func(t *testing.T) {
		g := NewWithT(t)

		getMountPropagationType = func(path string) (utils.MountPropagationType, error) {
			return utils.MountPropagationPrivate, nil
		}
		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
				Strict:     false,
			},
		}
		network := types.Network{
			Enabled:          ptr.To(true),
			LocalhostAddress: ptr.To("127.0.0.1"),
			PodCIDR:          ptr.To("192.0.2.0/24,2001:db8::/32"),
		}
		apiserver := types.APIServer{
			SecurePort: ptr.To(6443),
		}

		status, err := ApplyNetwork(context.Background(), snapM, apiserver, network, nil)

		g.Expect(err).To(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(Equal(fmt.Sprintf(networkDeployFailedMsgTmpl, err)))

		g.Expect(status.Version).To(Equal(CiliumAgentImageTag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(0))
	})
}

func validateNetworkValues(g Gomega, values map[string]any, network types.Network, snap snap.Snap) {

	ipv4CIDR, ipv6CIDR, err := utils.ParseCIDRs(network.GetPodCIDR())
	g.Expect(err).ToNot(HaveOccurred())

	bpfMount, err := utils.GetMountPath("bpf")
	g.Expect(err).ToNot(HaveOccurred())

	cgrMount, err := utils.GetMountPath("cgroup2")
	g.Expect(err).ToNot(HaveOccurred())

	if snap.Strict() {
		g.Expect(values["bpf"].(map[string]any)["root"]).To(Equal(bpfMount))
		g.Expect(values["cgroup"].(map[string]any)["hostRoot"]).To(Equal(cgrMount))
	}

	g.Expect(values["k8sServiceHost"]).To(Equal("127.0.0.1"))
	g.Expect(values["k8sServicePort"]).To(Equal(6443))
	g.Expect(values["ipam"].(map[string]any)["operator"].(map[string]any)["clusterPoolIPv4PodCIDRList"]).To(Equal(ipv4CIDR))
	g.Expect(values["ipam"].(map[string]any)["operator"].(map[string]any)["clusterPoolIPv6PodCIDRList"]).To(Equal(ipv6CIDR))
	g.Expect(values["ipv4"].(map[string]any)["enabled"]).To(Equal((ipv4CIDR != "")))
	g.Expect(values["ipv6"].(map[string]any)["enabled"]).To(Equal((ipv6CIDR != "")))
}
