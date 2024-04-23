package controllers_test

import (
	"context"
	"github.com/canonical/k8s/pkg/utils"
	"os"
	"path"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils/k8s"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpdateNodeConfigurationController(t *testing.T) {
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
		{
			name:            "ControlPlane_EmptyConfig",
			initialConfig:   types.ClusterConfig{},
			expectedConfig:  types.ClusterConfig{},
			expectedFailure: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			s := &mock.Snap{
				Mock: mock.Mock{
					EtcdPKIDir:                 path.Join(dir, "etcd-pki"),
					ServiceArgumentsDir:        path.Join(dir, "args"),
					UID:                        os.Getuid(),
					GID:                        os.Getgid(),
					KubernetesRESTClientGetter: genericclioptions.NewTestConfigFlags(),
				},
			}

			g := NewWithT(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			configProvider := &configProvider{config: tc.expectedConfig}
			kubeletConfigMap, err := tc.initialConfig.Kubelet.ToConfigMap(nil)
			g.Expect(err).ToNot(HaveOccurred())

			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "k8sd-config",
					Namespace: "kube-system",
				},
				Data: kubeletConfigMap,
			}
			clientset := fake.NewSimpleClientset(configMap)

			ctrl := controllers.NewUpdateNodeConfigurationController(s, func() {}, func() (*k8s.Client, error) {
				return &k8s.Client{Interface: clientset}, nil
			})
			go ctrl.Run(ctx, configProvider.getConfig)

			select {
			case ctrl.TriggerCh <- struct{}{}:
			case <-time.After(channelSendTimeout):
				g.Fail("Timed out while attempting to trigger controller reconcile loop")
			}

			select {
			case <-ctrl.ReconciledCh:
			case <-time.After(channelSendTimeout):
				g.Fail("Time out while waiting for the reconcile to complete")
			}

			result, err := clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "k8sd-config", metav1.GetOptions{})
			g.Expect(err).ToNot(HaveOccurred())
			expectedConfigMap, err := tc.expectedConfig.Kubelet.ToConfigMap(nil)
			g.Expect(err).ToNot(HaveOccurred())
			if tc.expectedFailure {
				g.Expect(result.Data).ToNot(Equal(expectedConfigMap))
			} else {
				g.Expect(result.Data).To(Equal(expectedConfigMap))
			}
		})
	}
}
