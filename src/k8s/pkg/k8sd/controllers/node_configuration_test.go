package controllers_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
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

	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	g.Expect(err).To(Not(HaveOccurred()))

	wrongPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	g.Expect(err).To(Not(HaveOccurred()))

	tests := []struct {
		name          string
		configmap     *corev1.ConfigMap
		expectArgs    map[string]string
		expectRestart bool
		privKey       *rsa.PrivateKey
		pubKey        *rsa.PublicKey
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
		{
			name: "WithSignature",
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
			privKey:       privKey,
			pubKey:        &privKey.PublicKey,
			expectRestart: true,
		},
		{
			name: "MissingPrivKey",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"cluster-dns":    "10.152.1.1",
					"cluster-domain": "test-cluster2.local",
					"cloud-provider": "provider",
				},
			},
			expectArgs: map[string]string{
				"--cluster-dns":    "10.152.1.1",
				"--cluster-domain": "test-cluster.local",
				"--cloud-provider": "provider",
			},
			pubKey:        &privKey.PublicKey,
			expectRestart: false,
		},
		{
			name: "InvalidSignature",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "k8sd-config", Namespace: "kube-system"},
				Data: map[string]string{
					"cluster-dns":    "10.152.1.1",
					"cluster-domain": "test-cluster2.local",
					"cloud-provider": "provider",
				},
			},
			expectArgs: map[string]string{
				"--cluster-dns":    "10.152.1.1",
				"--cluster-domain": "test-cluster.local",
				"--cloud-provider": "provider",
			},
			privKey:       wrongPrivKey,
			pubKey:        &privKey.PublicKey,
			expectRestart: false,
		},
	}

	clientset := fake.NewSimpleClientset()
	watcher := watch.NewFake()
	clientset.PrependWatchReactor("configmaps", k8stesting.DefaultWatchReactor(watcher, nil))

	s := &mock.Snap{
		Mock: mock.Mock{
			ServiceArgumentsDir:  filepath.Join(t.TempDir(), "args"),
			UID:                  os.Getuid(),
			GID:                  os.Getgid(),
			KubernetesNodeClient: &kubernetes.Client{Interface: clientset},
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())

	ctrl := controllers.NewNodeConfigurationController(s, func() {})

	keyCh := make(chan *rsa.PublicKey)

	go ctrl.Run(ctx, func(ctx context.Context) (*rsa.PublicKey, error) { return <-keyCh, nil })
	defer watcher.Stop()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s.RestartServiceCalledWith = nil

			g := NewWithT(t)

			if tc.privKey != nil {
				kubelet, err := types.KubeletFromConfigMap(tc.configmap.Data, nil)
				g.Expect(err).To(Not(HaveOccurred()))

				tc.configmap.Data, err = kubelet.ToConfigMap(tc.privKey)
				g.Expect(err).To(Not(HaveOccurred()))
			}

			watcher.Add(tc.configmap)

			keyCh <- tc.pubKey

			select {
			case <-ctrl.ReconciledCh():
			case <-time.After(channelSendTimeout):
				g.Fail("Time out while waiting for the reconcile to complete")
			}

			for ekey, evalue := range tc.expectArgs {
				val, err := snaputil.GetServiceArgument(s, "kubelet", ekey)
				g.Expect(err).To(Not(HaveOccurred()))
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
