package k8s

import (
	"context"
	"fmt"
	"log"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client with additional functionality
type Client struct {
	cacheLock        sync.RWMutex
	apiResourceCache map[string]*schema.GroupVersionResource

	clientset       kubernetes.Interface
	dynamicClient   dynamic.Interface
	discoveryClient *discovery.DiscoveryClient
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfig string, inCluster bool) (*Client, error) {
	config, err := buildConfig(kubeconfig, inCluster)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		clientset: clientset,
	}, nil
}

// buildConfig builds the Kubernetes configuration
func buildConfig(kubeconfig string, inCluster bool) (*rest.Config, error) {
	if inCluster {
		return rest.InClusterConfig()
	}

	if kubeconfig == "" {
		return nil, fmt.Errorf("kubeconfig path is required when not running in cluster")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func (c *Client) GetResource(ctx context.Context, kind, name, namespace string) (*unstructured.Unstructured, error) {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return nil, err
	}

	var obj *unstructured.Unstructured
	if namespace != "" {
		obj, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = c.dynamicClient.Resource(*gvr).Get(ctx, name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve resource: %w", err)
	}

	return obj, nil
}

func (c *Client) ListResources(ctx context.Context, kind, namespace, labelSelector string) (*unstructured.UnstructuredList, error) {
	gvr, err := c.getCachedGVR(kind)
	if err != nil {
		return nil, err
	}

	opts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	var objs *unstructured.UnstructuredList
	if namespace != "" {
		objs, err = c.dynamicClient.Resource(*gvr).Namespace(namespace).List(ctx, opts)
	} else {
		objs, err = c.dynamicClient.Resource(*gvr).List(ctx, opts)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	return objs, nil
}

func (c *Client) getCachedGVR(kind string) (*schema.GroupVersionResource, error) {
	c.cacheLock.RLock()
	if gvr, exists := c.apiResourceCache[kind]; exists {
		c.cacheLock.RUnlock()
		return gvr, nil
	}
	c.cacheLock.RUnlock()

	resourceLists, err := c.discoveryClient.ServerPreferredResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, fmt.Errorf("failed to retrieve API resources: %w", err)
	}

	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			log.Printf("Skipping invalid GroupVersion %s: %v\n", resourceList.GroupVersion, err)
			continue
		}
		for _, resource := range resourceList.APIResources {
			if resource.Kind == kind {
				gvr := &schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resource.Name,
				}
				c.cacheLock.Lock()
				c.apiResourceCache[kind] = gvr
				c.cacheLock.Unlock()
				return gvr, nil
			}
		}
	}

	return nil, fmt.Errorf("resource type %s not found", kind)
}
