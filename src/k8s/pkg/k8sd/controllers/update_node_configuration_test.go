package controllers_test

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils/k8s"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestUpdateNodeConfigurationController(t *testing.T) {
	testCases := []struct {
		name            string
		initialConfig   map[string]string
		expectedConfig  map[string]string
		expectedFailure bool
	}{
		{
			name: "ControlPlane_DefaultConfig",
			initialConfig: map[string]string{
				"test": "data",
			},
			expectedConfig: map[string]string{
				"ClusterDNS": "cluster.local",
			},
			expectedFailure: false,
		},
		{
			name: "ControlPlane_EmptyConfig",
			initialConfig: map[string]string{
				"test": "data",
			},
			expectedConfig:  nil, // Expecting empty ConfigMap data
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

			configProvider := &configProvider{}

			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "k8sd-config",
					Namespace: "kube-system",
				},
				Data: tc.initialConfig,
			}
			clientset := fake.NewSimpleClientset(configMap)

			if !tc.expectedFailure {
				clientset.PrependReactor("patch", "configmaps", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					configMap.Data = tc.expectedConfig
					return true, nil, nil
				})
			}

			ctrl := controllers.NewUpdateNodeConfigurationController(s, func() {}, func() (*k8s.Client, error) {
				return &k8s.Client{Interface: clientset}, nil
			})
			go ctrl.Run(ctx, configProvider.getConfig)

			select {
			case ctrl.UpdateCh <- struct{}{}:
			case <-time.After(channelSendTimeout):
				g.Fail("Timed out while attempting to trigger controller reconcile loop")
			}
			<-ctrl.ReconciledCh

			if tc.expectedFailure {
				g.Expect(configMap.Data).ToNot(Equal(tc.expectedConfig))
			} else {
				g.Expect(configMap.Data).To(Equal(tc.expectedConfig))
			}
		})
	}
}
