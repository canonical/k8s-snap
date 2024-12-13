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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestNodeConfigurationReconciler(t *testing.T) {
	testCases := []struct {
		name           string
		existingMap    bool
		expectedConfig types.ClusterConfig
	}{
		{
			name:        "ControlPlane_NotExist",
			existingMap: false,
			expectedConfig: types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDomain: utils.Pointer("cluster.local"),
				},
			},
		},
		{
			name:        "ControlPlane_ExistingConfig",
			existingMap: true,
			expectedConfig: types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDomain: utils.Pointer("cluster.local"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()

			// Setup scheme
			scheme := runtime.NewScheme()
			g.Expect(corev1.AddToScheme(scheme)).To(Succeed())

			// Setup objects for fake client
			objects := []client.Object{}
			if tc.existingMap {
				// Only create initial ConfigMap if test case requires it
				initialMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "k8sd-config",
						Namespace: "kube-system",
					},
					Data: map[string]string{"initial": "data"},
				}
				objects = append(objects, initialMap)
			}

			// Create fake client
			k8sClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(objects...).
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
				s,
				func() {}, // Mock ready function
			)

			reconciler.SetClient(k8sClient)
			reconciler.SetScheme(scheme)

			// Set the config getter
			reconciler.SetConfigGetter(func(context.Context) (types.ClusterConfig, error) {
				return tc.expectedConfig, nil
			})

			// Verify ConfigMap doesn't exist if it shouldn't
			if !tc.existingMap {
				var cm corev1.ConfigMap
				err := k8sClient.Get(ctx, client.ObjectKey{
					Name:      "k8sd-config",
					Namespace: "kube-system",
				}, &cm)
				g.Expect(err).To(HaveOccurred())
				g.Expect(apierrors.IsNotFound(err)).To(BeTrue(), "Expected ConfigMap not to exist")
			}

			// Trigger reconciliation
			req := reconcile.Request{
				NamespacedName: client.ObjectKey{
					Name:      "k8sd-config",
					Namespace: "kube-system",
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
			g.Expect(getResult.Data).To(Equal(expectedConfigMap))
		})
	}
}
