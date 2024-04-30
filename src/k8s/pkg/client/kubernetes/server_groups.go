package kubernetes

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListResourcesForGroupVersion lists the resources for a given group version (e.g. "cilium.io/v2alpha1")
func (c *Client) ListResourcesForGroupVersion(groupVersion string) (*v1.APIResourceList, error) {
	resources, err := c.Discovery().ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch server resources for %s: %w", groupVersion, err)
	}

	return resources, nil
}
