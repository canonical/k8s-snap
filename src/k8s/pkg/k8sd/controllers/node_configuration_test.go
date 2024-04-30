package controllers

import (
	"context"
	"crypto/rsa"
	"os"
	"path"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestConfigPropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := NewWithT(t)

	tests := []struct {
		name          string
		configmap     *corev1.ConfigMap
		expectArgs    map[string]string
		expectRestart bool
	}{
		{
			name: "Initial",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"cluster-dns":    "10.152.1.1",
					"cluster-domain": "test-cluster.local",
					"cloud-provider": "provider",
				},
			},
			expectArgs: map[string]string{
				"--cluster-dns":    "10.152.1.1",
				"--cluster-domain": "test-cluster.local",
				"--cloud-provider": "provider",
			},
			expectRestart: true,
		},
		{
			name: "IgnoreUnknownFields",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"non-existent-key1": "value1",
					"non-existent-key2": "value2",
					"non-existent-key3": "value3",
				},
			},
			expectArgs: map[string]string{
				"--cluster-dns":    "10.152.1.1",
				"--cluster-domain": "test-cluster.local",
				"--cloud-provider": "provider",
			},
		},
		{
			name: "RemoveClusterDNS",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"cluster-dns": "",
				},
			},
			expectArgs: map[string]string{
				"--cluster-dns":    "",
				"--cluster-domain": "test-cluster.local",
				"--cloud-provider": "provider",
			},
			expectRestart: true,
		},
		{
			name: "UpdateDNS",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"cluster-domain": "test-cluster2.local",
					"cluster-dns":    "10.152.1.3",
				},
			},
			expectArgs: map[string]string{
				"--cluster-domain": "test-cluster2.local",
				"--cluster-dns":    "10.152.1.3",
				"--cloud-provider": "provider",
			},
			expectRestart: true,
		},
		{
			name: "PreserveClusterDomain",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"cluster-dns": "10.152.1.3",
				},
			},
			expectArgs: map[string]string{
				"--cluster-domain": "test-cluster2.local",
				"--cluster-dns":    "10.152.1.3",
				"--cloud-provider": "provider",
			},
		},
	}

	clientset := fake.NewSimpleClientset()
	watcher := watch.NewFake()
	clientset.PrependWatchReactor("configmaps", k8stesting.DefaultWatchReactor(watcher, nil))

	s := &mock.Snap{
		Mock: mock.Mock{
			ServiceArgumentsDir:  path.Join(t.TempDir(), "args"),
			UID:                  os.Getuid(),
			GID:                  os.Getgid(),
			KubernetesNodeClient: &kubernetes.Client{Interface: clientset},
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())

	ctrl := NewNodeConfigurationController(s, func() {})

	// TODO: add test with signing key
	go ctrl.Run(ctx, func(ctx context.Context) (*rsa.PublicKey, error) { return nil, nil })
	defer watcher.Stop()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s.RestartServiceCalledWith = nil

			g := NewWithT(t)

			watcher.Add(tc.configmap)

			// TODO: this is to ensure that the controller has handled the event. This should ideally
			// be replaced with something like a "<-sentCh" instead
			time.Sleep(100 * time.Millisecond)

			for ekey, evalue := range tc.expectArgs {
				val, err := snaputil.GetServiceArgument(s, "kubelet", ekey)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(evalue))
			}

			if tc.expectRestart {
				g.Expect(s.RestartServiceCalledWith).To(Equal([]string{"kubelet"}))
			} else {
				g.Expect(s.RestartServiceCalledWith).To(BeEmpty())
			}
		})
	}
}
