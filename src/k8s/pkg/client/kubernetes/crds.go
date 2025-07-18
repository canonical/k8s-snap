package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils/control"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// TODO(ben): Add unittests.
// ApplyCRDs applies all CRD YAML files in the specified directory.
func (c *Client) ApplyCRDs(ctx context.Context, crdsDir string) error {
	log := log.FromContext(ctx).WithValues("kubernetes", "ApplyCRDs", "dir", crdsDir)

	files, err := os.ReadDir(crdsDir)
	if err != nil {
		return fmt.Errorf("failed to read CRD directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".yaml" {
			continue // Skip directories and non-YAML files
		}

		crdPath := filepath.Join(crdsDir, file.Name())

		if err := c.ApplyCRD(ctx, crdPath); err != nil {
			return fmt.Errorf("failed to apply CRD %s: %w", file.Name(), err)
		}
	}

	log.Info("Successfully applied CRDs.", "numOfFiles", len(files))
	return nil
}

// TODO(ben): Add unittests.
// ApplyCRD reads and applies a single CRD YAML file.
func (c *Client) ApplyCRD(ctx context.Context, filePath string) error {
	log := log.FromContext(ctx).WithValues("kubernetes", "ApplyCRD", "file", filePath)

	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Create API Extensions Client for managing CRDs
	apiExtClient, err := apiextensionsclient.NewForConfig(c.RESTConfig())
	if err != nil {
		return fmt.Errorf("failed to create API extensions client: %w", err)
	}

	// Decode YAML into an unstructured object
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	if _, _, err := dec.Decode(yamlFile, nil, obj); err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}

	// Convert unstructured object to a CRD
	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := c.convertUnstructuredToCRD(obj, crd); err != nil {
		return fmt.Errorf("failed to convert to CRD: %w", err)
	}

	// TODO(ben): Consider using `Apply` instead.
	existing, err := apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, v1.GetOptions{})
	if err == nil {
		// CRD exists, update it
		crd.ResourceVersion = existing.ResourceVersion
		_, err = apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, v1.UpdateOptions{})
	} else if apierrors.IsNotFound(err) {
		// CRD doesn't exist, create it
		_, err = apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, v1.CreateOptions{})
	}
	if err != nil {
		return fmt.Errorf("failed to apply CRD: %w", err)
	}

	log.V(1).Info("Applied CRD", "name", crd.Name, "version", crd.APIVersion, "kind", crd.Kind)

	if waitErr := control.WaitUntilReady(ctx, func() (bool, error) {
		return c.CRDExists(ctx, crd.Name)
	}); waitErr != nil {
		return fmt.Errorf("failed to wait for CRD to be ready: %w", waitErr)
	}

	log.Info("CRD is now available", "name", crd.Name)
	return nil
}

// CRDExists checks if a given CRD exists in the cluster.
func (c *Client) CRDExists(ctx context.Context, crdName string) (bool, error) {
	apiExtClient, err := apiextensionsclient.NewForConfig(c.RESTConfig())
	if err != nil {
		return false, fmt.Errorf("failed to create API extensions client: %w", err)
	}

	_, err = apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, v1.GetOptions{})
	if err == nil {
		return true, nil // CRD exists
	}

	if apierrors.IsNotFound(err) {
		return false, nil // CRD does not exist
	}

	return false, fmt.Errorf("failed to check CRD existence: %w", err)
}

// convertUnstructuredToCRD converts an unstructured object to a CRD object.
func (c *Client) convertUnstructuredToCRD(obj *unstructured.Unstructured, crd *apiextensionsv1.CustomResourceDefinition) error {
	crdBytes, err := obj.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal unstructured object: %w", err)
	}

	if err := json.Unmarshal(crdBytes, crd); err != nil {
		return fmt.Errorf("failed to unmarshal to CRD: %w", err)
	}
	return nil
}
