package metrics_server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/canonical/k8s/pkg/client/helm"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	metrics_server "github.com/canonical/k8s/pkg/k8sd/features/metrics-server"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestApplyMetricsServer(t *testing.T) {

	helmErr := errors.New("failed to apply")
	for _, tc := range []struct {
		name        string
		config      types.MetricsServer
		expectState helm.State
		helmError   error
	}{
		{
			name: "EnableWithoutHelmError",
			config: types.MetricsServer{
				Enabled: utils.Pointer(true),
			},
			expectState: helm.StatePresent,
			helmError:   nil,
		},
		{
			name: "DisableWithoutHelmError",
			config: types.MetricsServer{
				Enabled: utils.Pointer(false),
			},
			expectState: helm.StateDeleted,
			helmError:   nil,
		},
		{
			name: "EnableWithHelmError",
			config: types.MetricsServer{
				Enabled: utils.Pointer(true),
			},
			expectState: helm.StatePresent,
			helmError:   helmErr,
		},
		{
			name: "DisableWithHelmError",
			config: types.MetricsServer{
				Enabled: utils.Pointer(false),
			},
			expectState: helm.StateDeleted,
			helmError:   helmErr,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			h := &helmmock.Mock{
				ApplyErr: tc.helmError,
			}
			s := &snapmock.Snap{
				Mock: snapmock.Mock{
					HelmClient: h,
				},
			}

			status, err := metrics_server.ApplyMetricsServer(context.Background(), s, tc.config, nil)
			if tc.helmError == nil {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err).To(HaveOccurred())
			}

			g.Expect(h.ApplyCalledWith).To(ConsistOf(SatisfyAll(
				HaveField("Chart.Name", Equal("metrics-server")),
				HaveField("Chart.Namespace", Equal("kube-system")),
				HaveField("State", Equal(tc.expectState)),
			)))
			if errors.Is(tc.helmError, helmErr) {
				g.Expect(status.Message).To(ContainSubstring(helmErr.Error()))
			} else if tc.config.GetEnabled() {
				g.Expect(status.Message).To(Equal("enabled"))
			} else {
				g.Expect(status.Message).To(Equal("disabled"))
			}
		})
	}

	t.Run("Annotations", func(t *testing.T) {
		g := NewWithT(t)
		h := &helmmock.Mock{}
		s := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: h,
			},
		}

		cfg := types.MetricsServer{
			Enabled: utils.Pointer(true),
		}
		annotations := types.Annotations{
			"k8sd/v1alpha1/metrics-server/image-repo": "custom-image",
			"k8sd/v1alpha1/metrics-server/image-tag":  "custom-tag",
		}

		status, err := metrics_server.ApplyMetricsServer(context.Background(), s, cfg, annotations)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(h.ApplyCalledWith).To(ConsistOf(HaveField("Values", HaveKeyWithValue("image", SatisfyAll(
			HaveKeyWithValue("repository", "custom-image"),
			HaveKeyWithValue("tag", "custom-tag"),
		)))))
		g.Expect(status.Message).To(Equal("enabled"))
	})
}
