package local_storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/client/helm/loader"
	helmmock "github.com/canonical/k8s/pkg/client/helm/mock"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/localpv"
	localpv_local_storage "github.com/canonical/k8s/pkg/k8sd/features/localpv/local-storage"
	"github.com/canonical/k8s/pkg/k8sd/types"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
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
		cfg := types.ClusterConfig{
			LocalStorage: types.LocalStorage{
				Enabled:       ptr.To(false),
				Default:       ptr.To(true),
				ReclaimPolicy: ptr.To("reclaim-policy"),
				LocalPath:     ptr.To("local-path"),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&localpv.ChartFS))

		base := features.NewReconciler(localpv_local_storage.Manifest, snapM, mc, nil, func() {})
		reconciler := localpv_local_storage.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Version).To(Equal(localpv_local_storage.Manifest.GetImage(localpv_local_storage.RawFileImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(localpv_local_storage.Manifest.GetChart(localpv_local_storage.RawFileChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StateDeleted))

		validateValues(g, callArgs.Values, cfg.LocalStorage)
	})
	t.Run("Success", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		cfg := types.ClusterConfig{
			LocalStorage: types.LocalStorage{
				Enabled:       ptr.To(false),
				Default:       ptr.To(true),
				ReclaimPolicy: ptr.To("reclaim-policy"),
				LocalPath:     ptr.To("local-path"),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&localpv.ChartFS))

		base := features.NewReconciler(localpv_local_storage.Manifest, snapM, mc, nil, func() {})
		reconciler := localpv_local_storage.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Version).To(Equal(localpv_local_storage.Manifest.GetImage(localpv_local_storage.RawFileImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(localpv_local_storage.Manifest.GetChart(localpv_local_storage.RawFileChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StateDeleted))

		validateValues(g, callArgs.Values, cfg.LocalStorage)
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
		cfg := types.ClusterConfig{
			LocalStorage: types.LocalStorage{
				Enabled:       ptr.To(true),
				Default:       ptr.To(true),
				ReclaimPolicy: ptr.To("reclaim-policy"),
				LocalPath:     ptr.To("local-path"),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&localpv.ChartFS))

		base := features.NewReconciler(localpv_local_storage.Manifest, snapM, mc, nil, func() {})
		reconciler := localpv_local_storage.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).To(MatchError(applyErr))
		g.Expect(status.Enabled).To(BeFalse())
		g.Expect(status.Message).To(ContainSubstring(applyErr.Error()))
		g.Expect(status.Version).To(Equal(localpv_local_storage.Manifest.GetImage(localpv_local_storage.RawFileImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(localpv_local_storage.Manifest.GetChart(localpv_local_storage.RawFileChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))

		validateValues(g, callArgs.Values, cfg.LocalStorage)
	})
	t.Run("Success", func(t *testing.T) {
		g := NewWithT(t)

		helmM := &helmmock.Mock{}
		snapM := &snapmock.Snap{
			Mock: snapmock.Mock{
				HelmClient: helmM,
			},
		}
		cfg := types.ClusterConfig{
			LocalStorage: types.LocalStorage{
				Enabled:       ptr.To(true),
				Default:       ptr.To(true),
				ReclaimPolicy: ptr.To("reclaim-policy"),
				LocalPath:     ptr.To("local-path"),
			},
		}

		mc := snapM.HelmClient(loader.NewEmbedLoader(&localpv.ChartFS))

		base := features.NewReconciler(localpv_local_storage.Manifest, snapM, mc, nil, func() {})
		reconciler := localpv_local_storage.NewReconciler(base)

		status, err := reconciler.Reconcile(context.Background(), cfg)

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.Enabled).To(BeTrue())
		g.Expect(status.Version).To(Equal(localpv_local_storage.Manifest.GetImage(localpv_local_storage.RawFileImageName).Tag))
		g.Expect(helmM.ApplyCalledWith).To(HaveLen(1))

		callArgs := helmM.ApplyCalledWith[0]
		g.Expect(callArgs.Chart).To(Equal(localpv_local_storage.Manifest.GetChart(localpv_local_storage.RawFileChartName)))
		g.Expect(callArgs.State).To(Equal(helm.StatePresent))

		validateValues(g, callArgs.Values, cfg.LocalStorage)
	})
}

func validateValues(g Gomega, values map[string]any, cfg types.LocalStorage) {
	sc := values["storageClass"].(map[string]any)
	g.Expect(sc["isDefault"]).To(Equal(cfg.GetDefault()))
	g.Expect(sc["reclaimPolicy"]).To(Equal(cfg.GetReclaimPolicy()))

	storage := values["node"].(map[string]any)["storage"].(map[string]any)
	g.Expect(storage["path"]).To(Equal(cfg.GetLocalPath()))
}
