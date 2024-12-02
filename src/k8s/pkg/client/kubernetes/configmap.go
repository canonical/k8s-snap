package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func (c *Client) WatchConfigMap(ctx context.Context, namespace string, name string, reconcile func(configMap *v1.ConfigMap) error) error {
	log := log.FromContext(ctx).WithValues("namespace", namespace, "name", name)
	w, err := c.CoreV1().ConfigMaps(namespace).Watch(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	if err != nil {
		return fmt.Errorf("failed to watch configmap namespace=%s name=%s: %w", namespace, name, err)
	}
	defer w.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-w.ResultChan():
			if !ok {
				return fmt.Errorf("watch closed")
			}
			configMap, ok := evt.Object.(*v1.ConfigMap)
			if !ok {
				return fmt.Errorf("expected a ConfigMap but received %#v", evt.Object)
			}

			if err := reconcile(configMap); err != nil {
				log.Error(err, "Reconcile ConfigMap failed")
			}
		}
	}
}

func (c *Client) UpdateConfigMap(ctx context.Context, namespace string, name string, data map[string]string) (*v1.ConfigMap, error) {
	opts := applyv1.ConfigMap(name, namespace).WithData(data)
	configmap, err := c.CoreV1().ConfigMaps(namespace).Apply(ctx, opts, metav1.ApplyOptions{FieldManager: "ck-k8s-client"})
	if err != nil {
		return nil, fmt.Errorf("failed to update configmap, namespace: %s name: %s: %w", namespace, name, err)
	}
	return configmap, nil
}

// ForceConfigMapReconcilation adds an annotation to the ConfigMap to force a reconciliation.
func (c *Client) ForceConfigMapReconcilation(ctx context.Context, namespace string, configMapName string) error {
	// Fetch the current ConfigMap
	configMap, err := c.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap %s: %w", configMapName, err)
	}

	// Add or update the annotation
	if configMap.Annotations == nil {
		configMap.Annotations = make(map[string]string)
	}
	configMap.Annotations["reconcile-time"] = time.Now().Format(time.RFC3339)

	// Update the ConfigMap with the new annotation
	_, err = c.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ConfigMap %s: %w", configMapName, err)
	}

	return nil
}
