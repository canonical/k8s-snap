package controllers_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestNodeConfigurationReconciler(t *testing.T) {
	testCases := []struct {
		name            string
		initialConfig   types.ClusterConfig
		expectedConfig  types.ClusterConfig
		expectedFailure bool
	}{
		{
			name:          "ControlPlane_DefaultConfig",
			initialConfig: types.ClusterConfig{},
			expectedConfig: types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDomain: utils.Pointer("cluster.local"),
				},
			},
			expectedFailure: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()

			// Setup scheme
			scheme := runtime.NewScheme()
			g.Expect(corev1.AddToScheme(scheme)).To(Succeed())

			// Create initial ConfigMap
			kubeletConfigMap, err := tc.initialConfig.Kubelet.ToConfigMap(nil)
			g.Expect(err).ToNot(HaveOccurred())

			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "k8sd-config",
					Namespace: "kube-systems",
				},
				Data: kubeletConfigMap,
			}

			// Create fake client
			k8sClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(configMap).
				Build()

			// Setup mock snap
			s := &mock.Snap{
				Mock: mock.Mock{
					EtcdPKIDir:          filepath.Join(t.TempDir(), "etcd-pki"),
					ServiceArgumentsDir: filepath.Join(t.TempDir(), "args"),
				},
			}

			// Create controller
			reconciler := controllers.NewNodeConfigurationReconciler(
				k8sClient,
				scheme,
				s,
				func() {}, // Mock ready function
			)

			// Set the config getter
			reconciler.SetConfigGetter(func(context.Context) (types.ClusterConfig, error) {
				return tc.expectedConfig, nil
			})

			// Trigger reconciliation
			req := reconcile.Request{
				NamespacedName: client.ObjectKey{
					Name:      "k8sd-config",
					Namespace: "kube-systems",
				},
			}

			result, err := reconciler.Reconcile(ctx, req)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(reconcile.Result{}))

			// Verify results
			var getResult corev1.ConfigMap
			g.Expect(k8sClient.Get(ctx, client.ObjectKey{
				Name:      "k8sd-config",
				Namespace: "kube-system",
			}, &getResult)).To(Succeed())

			expectedConfigMap, err := tc.expectedConfig.Kubelet.ToConfigMap(nil)
			g.Expect(err).ToNot(HaveOccurred())

			if tc.expectedFailure {
				g.Expect(getResult.Data).ToNot(Equal(expectedConfigMap))
			} else {
				g.Expect(getResult.Data).To(Equal(expectedConfigMap))
			}
		})
	}
}
