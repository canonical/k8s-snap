package k8s

import (
	"context"

	policy "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EvictPod evicts a pod from a namespace.
func EvictPod(ctx context.Context, client *k8sClient, namespace string, name string) error {
	return client.PolicyV1().Evictions(namespace).Evict(ctx, &policy.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		}})
}
