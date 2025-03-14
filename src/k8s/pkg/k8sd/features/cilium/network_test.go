package cilium_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/cilium"
	"github.com/canonical/k8s/pkg/client/helm"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"
	testenv "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
	"k8s.io/utils/ptr"
)

// NOTE(hue): status.Message is not checked sometimes to avoid unnecessary complexity

var annotations = types.Annotations{
	apiv1_annotations.AnnotationDevices:             "eth+ lxdbr+",
	apiv1_annotations.AnnotationDirectRoutingDevice: "eth0",
}

func TestNetworkDisabled(t *testing.T) {
	testenv.WithState(t, func(ctx context.Context, s state.State) {
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

			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, nil)

			g.Expect(err).To(MatchError(applyErr))
			g.Expect(status.Enabled).To(BeFalse())
			g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.NetworkDeleteFailedMsgTmpl, err)))
			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

			callArgs := helmM.ApplyCalledWith[0]
			g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
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

			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, nil)

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
	})
}

func TestNetworkEnabled(t *testing.T) {
	testenv.WithState(t, func(ctx context.Context, s state.State) {
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

			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, nil)

			g.Expect(err).To(HaveOccurred())
			g.Expect(status.Enabled).To(BeFalse())
			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(BeEmpty())
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
				Enabled: ptr.To(true),
				PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
			}
			apiserver := types.APIServer{
				SecurePort: ptr.To(6443),
			}

			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, annotations)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.Enabled).To(BeTrue())
			g.Expect(status.Message).To(Equal(cilium.EnabledMsg))
			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

			callArgs := helmM.ApplyCalledWith[0]
			g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
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
				Enabled: ptr.To(true),
				PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
			}
			apiserver := types.APIServer{
				SecurePort: ptr.To(6443),
			}

			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, annotations)

			g.Expect(err).To(MatchError(applyErr))
			g.Expect(status.Enabled).To(BeFalse())
			g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.NetworkDeployFailedMsgTmpl, err)))
			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

			callArgs := helmM.ApplyCalledWith[0]
			g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
			g.Expect(callArgs.State).To(Equal(helm.StatePresent))
			validateNetworkValues(g, callArgs.Values, network, snapM)
		})

		t.Run("CNIExclusive", func(t *testing.T) {
			g := NewWithT(t)

			helmM := &helmmock.Mock{}
			snapM := &snapmock.Snap{
				Mock: snapmock.Mock{
					HelmClient: helmM,
				},
			}
			network := types.Network{
				Enabled: ptr.To(true),
				PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
			}
			apiserver := types.APIServer{
				SecurePort: ptr.To(6443),
			}

			testAnnotations := types.Annotations{
				apiv1_annotations.AnnotationDevices:             "eth+ lxdbr+",
				apiv1_annotations.AnnotationDirectRoutingDevice: "eth0",
				apiv1_annotations.AnnotationCNIExclusive:        "true",
			}
			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, testAnnotations)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.Enabled).To(BeTrue())
			g.Expect(status.Message).To(Equal(cilium.EnabledMsg))
			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

			callArgs := helmM.ApplyCalledWith[0]
			g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
			g.Expect(callArgs.State).To(Equal(helm.StatePresent))

			cniValues := callArgs.Values["cni"].(map[string]interface{})
			g.Expect(cniValues["exclusive"]).To(BeTrue())
		})

		t.Run("SCTPEnabled", func(t *testing.T) {
			g := NewWithT(t)

			helmM := &helmmock.Mock{}
			snapM := &snapmock.Snap{
				Mock: snapmock.Mock{
					HelmClient: helmM,
				},
			}
			network := types.Network{
				Enabled: ptr.To(true),
				PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
			}
			apiserver := types.APIServer{
				SecurePort: ptr.To(6443),
			}

			testAnnotations := types.Annotations{
				apiv1_annotations.AnnotationDevices:             "eth+ lxdbr+",
				apiv1_annotations.AnnotationDirectRoutingDevice: "eth0",
				apiv1_annotations.AnnotationSCTPEnabled:         "true",
			}
			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, testAnnotations)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(status.Enabled).To(BeTrue())
			g.Expect(status.Message).To(Equal(cilium.EnabledMsg))
			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

			callArgs := helmM.ApplyCalledWith[0]
			g.Expect(callArgs.Chart).To(Equal(cilium.ChartCilium))
			g.Expect(callArgs.State).To(Equal(helm.StatePresent))

			sctpValues := callArgs.Values["sctp"].(map[string]interface{})
			g.Expect(sctpValues["enabled"]).To(BeTrue())
		})
	})
}

func TestNetworkMountPath(t *testing.T) {
	testenv.WithState(t, func(ctx context.Context, s state.State) {
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
					Enabled: ptr.To(true),
					PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
				}
				apiserver := types.APIServer{
					SecurePort: ptr.To(6443),
				}
				cilium.GetMountPath = func(fsType string) (string, error) {
					if fsType == tc.name {
						return "", mountPathErr
					}
					return tc.name, nil
				}

				status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, nil)

				g.Expect(err).To(HaveOccurred())
				g.Expect(err).To(MatchError(mountPathErr))
				g.Expect(status.Enabled).To(BeFalse())
				g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.NetworkDeployFailedMsgTmpl, err)))
				g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
				g.Expect(helmM.ApplyCalledWith).To(BeEmpty())
			})
		}
	})
}

func TestNetworkMountPropagationType(t *testing.T) {
	testenv.WithState(t, func(ctx context.Context, s state.State) {
		t.Run("failedGetMountSys", func(t *testing.T) {
			g := NewWithT(t)

			mountErr := errors.New("/sys not found")
			cilium.GetMountPropagationType = func(path string) (utils.MountPropagationType, error) {
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
				Enabled: ptr.To(true),
				PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
			}
			apiserver := types.APIServer{
				SecurePort: ptr.To(6443),
			}

			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, nil)

			g.Expect(err).To(HaveOccurred())
			g.Expect(err).To(MatchError(mountErr))
			g.Expect(status.Enabled).To(BeFalse())
			g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.NetworkDeployFailedMsgTmpl, err)))

			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(BeEmpty())
		})

		t.Run("MountPropagationPrivateOnLXDError", func(t *testing.T) {
			g := NewWithT(t)

			cilium.GetMountPropagationType = func(path string) (utils.MountPropagationType, error) {
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
				Enabled: ptr.To(true),
				PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
			}
			apiserver := types.APIServer{
				SecurePort: ptr.To(6443),
			}
			logger := ktesting.NewLogger(t, ktesting.NewConfig(ktesting.BufferLogs(true)))
			ctx := klog.NewContext(context.Background(), logger)

			status, err := cilium.ApplyNetwork(ctx, snapM, s, apiserver, network, nil)

			g.Expect(err).To(HaveOccurred())
			g.Expect(status.Enabled).To(BeFalse())
			g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.NetworkDeployFailedMsgTmpl, err)))

			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(BeEmpty())
			testingLogger, ok := logger.GetSink().(ktesting.Underlier)
			if !ok {
				panic("Should have had a ktesting LogSink!?")
			}
			g.Expect(testingLogger.GetBuffer().String()).To(ContainSubstring("Failed to check if running on LXD"))
		})

		t.Run("MountPropagationPrivateOnLXD", func(t *testing.T) {
			g := NewWithT(t)

			cilium.GetMountPropagationType = func(path string) (utils.MountPropagationType, error) {
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
				Enabled: ptr.To(true),
				PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
			}
			apiserver := types.APIServer{
				SecurePort: ptr.To(6443),
			}

			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, nil)

			g.Expect(err).To(HaveOccurred())
			g.Expect(status.Enabled).To(BeFalse())
			g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.NetworkDeployFailedMsgTmpl, err)))

			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(BeEmpty())
		})

		t.Run("MountPropagationPrivate", func(t *testing.T) {
			g := NewWithT(t)

			cilium.GetMountPropagationType = func(_ string) (utils.MountPropagationType, error) {
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
				Enabled: ptr.To(true),
				PodCIDR: ptr.To("192.0.2.0/24,2001:db8::/32"),
			}
			apiserver := types.APIServer{
				SecurePort: ptr.To(6443),
			}

			status, err := cilium.ApplyNetwork(context.Background(), snapM, s, apiserver, network, nil)

			g.Expect(err).To(HaveOccurred())
			g.Expect(status.Enabled).To(BeFalse())
			g.Expect(status.Message).To(Equal(fmt.Sprintf(cilium.NetworkDeployFailedMsgTmpl, err)))

			g.Expect(status.Version).To(Equal(cilium.CiliumAgentImageTag))
			g.Expect(helmM.ApplyCalledWith).To(BeEmpty())
		})
	})
}

func validateNetworkValues(g Gomega, values map[string]any, network types.Network, snap snap.Snap) {
	ipv4CIDR, ipv6CIDR, err := utils.SplitCIDRStrings(network.GetPodCIDR())
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

	devices, exists := annotations.Get(apiv1_annotations.AnnotationDevices)
	if exists {
		g.Expect(values["devices"]).To(Equal(devices))
	}

	directRoutingDevice, exists := annotations.Get(apiv1_annotations.AnnotationDirectRoutingDevice)
	if exists {
		g.Expect(values["nodePort"].(map[string]any)["directRoutingDevice"]).To(Equal(directRoutingDevice))
	}

	_, exists = annotations.Get(apiv1_annotations.AnnotationCNIExclusive)
	cniValues := values["cni"].(map[string]interface{})
	g.Expect(cniValues["exclusive"]).To(Equal(exists))

	_, exists = annotations.Get(apiv1_annotations.AnnotationSCTPEnabled)
	sctpValues := values["sctp"].(map[string]interface{})
	g.Expect(sctpValues["enabled"]).To(Equal(exists))
}
