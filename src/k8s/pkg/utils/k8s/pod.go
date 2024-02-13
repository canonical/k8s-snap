package k8s

import (
	"context"

	policy "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EvictPod evicts a pod from a namespace.
func (c *Client) EvictPod(ctx context.Context, namespace string, name string) error {
	return c.PolicyV1().Evictions(namespace).Evict(ctx, &policy.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		}})
}
