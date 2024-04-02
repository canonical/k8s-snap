package controllers

import (
	"context"
	"net"
	"os"
	"path"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/k8s"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestConfigPropagation(t *testing.T) {
	ctx := context.Background()

	g := NewWithT(t)

	dir := t.TempDir()

	s := &mock.Snap{
		Mock: mock.Mock{
			KubernetesPKIDir:    path.Join(dir, "pki"),
			KubernetesConfigDir: path.Join(dir, "k8s-config"),
			KubeletRootDir:      path.Join(dir, "kubelet-root"),
			ServiceArgumentsDir: path.Join(dir, "args"),
			ContainerdSocketDir: path.Join(dir, "containerd-run"),
			OnLXD:               false,
			UID:                 os.Getuid(),
			GID:                 os.Getgid(),
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())

	// Call the kubelet control plane setup function
	g.Expect(setup.KubeletControlPlane(s, "dev", net.ParseIP("192.168.0.1"), "10.152.1.1", "test-cluster.local", "provider")).To(Succeed())

	tests := []struct {
		name            string
		configmap       *corev1.ConfigMap
		expectedUpdates map[string]string
	}{
		{
			name: "ignore non-existent keys",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"non-existent-key1": "value1",
					"non-existent-key2": "value2",
					"non-existent-key3": "value3",
				},
			},
		},
		{
			name: "remove cluster-dns on missing key",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data:       map[string]string{},
			},
			expectedUpdates: map[string]string{
				"--cluster-dns": "",
			},
		},
		{
			name: "update node configuration",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"cluster-domain": "test-cluster2.local",
					"cluster-dns":    "10.152.1.3",
				},
			},
			expectedUpdates: map[string]string{
				"--cluster-domain": "test-cluster2.local",
				"--cluster-dns":    "10.152.1.3",
			},
		},
		{
			name: "cluster-domain remains on missing key",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"cluster-dns": "10.152.1.3",
				},
			},
			expectedUpdates: map[string]string{
				"--cluster-domain": "cluster.local",
				"--cluster-dns":    "10.152.1.3",
			},
		},
	}

	clientset := fake.NewSimpleClientset()
	watcher := watch.NewFake()
	clientset.PrependWatchReactor("configmaps", k8stesting.DefaultWatchReactor(watcher, nil))

	configController := NewNodeConfigurationController(s, func(ctx context.Context) *k8s.Client {
		return &k8s.Client{Interface: clientset}
	})

	go configController.Run(ctx)

	defer watcher.Stop()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			watcher.Add(tc.configmap)
			time.Sleep(100 * time.Millisecond)

			for ekey, evalue := range tc.expectedUpdates {
				val, err := snaputil.GetServiceArgument(s, "kubelet", ekey)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(val).To(Equal(evalue))
			}
		})
	}
}
