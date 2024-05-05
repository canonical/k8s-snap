package kubernetes

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestWatchConfigMap(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		configmap *corev1.ConfigMap
	}{
		{
			name:      "pass nil object",
			configmap: nil,
		},
		{
			name: "example configmap with values",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "test-config", Namespace: "kube-system"},
				Data: map[string]string{
					"non-existent-key1": "value1",
					"non-existent-key2": "value2",
					"non-existent-key3": "value3",
				},
			},
		},
		{
			name: "configmap with empty data",
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "test-config", Namespace: "kube-system"},
				Data:       map[string]string{},
			},
		},
	}

	clientset := fake.NewSimpleClientset()
	watcher := watch.NewFake()
	clientset.PrependWatchReactor("configmaps", k8stesting.DefaultWatchReactor(watcher, nil))

	client := &Client{Interface: clientset}

	doneCh := make(chan *corev1.ConfigMap)

	go client.WatchConfigMap(ctx, "kube-system", "test-config", func(configMap *corev1.ConfigMap) error {
		doneCh <- configMap
		if configMap == nil {
			return fmt.Errorf("unexpected nil map test case error")
		}
		return nil
	})

	defer watcher.Stop()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			watcher.Add(tc.configmap)
			select {
			case recv := <-doneCh:
				if tc.configmap != nil {
					g.Expect(recv.Data).To(Equal(tc.configmap.Data))
					g.Expect(recv.Name).To(Equal(tc.configmap.Name))
					g.Expect(recv.Namespace).To(Equal(tc.configmap.Namespace))
				}
			case <-time.After(time.Second):
				t.Fatal("Timed out waiting for watch to complete")
			}

		})
	}
}

func TestUpdateConfigMap(t *testing.T) {
	ctx := context.Background()

	g := NewWithT(t)

	existingObjs := []runtime.Object{
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "test-config", Namespace: "kube-system"},
			Data: map[string]string{
				"existing-key": "old-value",
			},
		},
	}

	clientset := fake.NewSimpleClientset(existingObjs...)
	client := &Client{Interface: clientset}

	updateData := map[string]string{
		"existing-key": "change-value",
		"new-key":      "new-value",
	}
	cm, err := client.UpdateConfigMap(ctx, "kube-system", "test-config", updateData)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Data).To(Equal(updateData))
}
