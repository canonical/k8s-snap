package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/canonical/k8s/pkg/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// TODO: If the upgrade CRD grows, consider using kubebuilder.
type Upgrade struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Status struct {
		Phase         string   `json:"phase"`
		UpgradedNodes []string `json:"upgradedNodes"`
	} `json:"status"`
}

func (c *Client) K8sdIoRestClient() (*rest.RESTClient, error) {
	k8sdConfig := c.RESTConfig()
	k8sdConfig.GroupVersion = &schema.GroupVersion{Group: "k8sd.io", Version: "v1alpha"}
	k8sdConfig.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())

	restClient, err := rest.RESTClientFor(k8sdConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}
	return restClient, nil
}

func (c *Client) GetInProgressUpgrade(ctx context.Context) (*Upgrade, error) {
	log := log.FromContext(ctx).WithValues("upgrades", "GetInProgressUpgrade")

	restClient, err := c.K8sdIoRestClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}

	upgrades, err := restClient.Get().AbsPath("/apis/k8sd.io/v1alpha/upgrades").DoRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get upgrades: %w", err)
	}
	log.Info("Got upgrades", "upgrades", string(upgrades))
	var result struct {
		Items []Upgrade `json:"items"`
	}
	if err := json.Unmarshal(upgrades, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upgrades: %w", err)
	}

	for _, upgrade := range result.Items {
		if upgrade.Status.Phase != "Failed" && upgrade.Status.Phase != "Completed" {
			return &upgrade, nil
		}
	}
	return nil, nil
}

// GetUpgradedNodes returns the list of upgraded nodes or an error.
func (c *Client) GetUpgradedNodes(ctx context.Context) ([]string, error) {
	upgrade, err := c.GetInProgressUpgrade(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get in-progress upgrade: %w", err)
	}

	if upgrade == nil {
		return nil, fmt.Errorf("no upgrade in progress")
	}

	return upgrade.Status.UpgradedNodes, nil
}

// MarkNodeUpgradeDone marks the current node as upgraded in the Upgrade CRD.
func (c *Client) MarkNodeUpgradeDone(ctx context.Context, nodeName string) error {
	log := log.FromContext(ctx).WithValues("upgrades", "markNodeUpgradeDone", "node", nodeName)

	upgrade, err := c.GetInProgressUpgrade(ctx)
	if err != nil {
		return fmt.Errorf("failed to get in-progress upgrade: %w", err)
	}

	if upgrade == nil {
		return fmt.Errorf("no upgrade in progress")
	}

	if slices.Contains(upgrade.Status.UpgradedNodes, nodeName) {
		log.Info("Node is already marked as upgraded", "node", nodeName)
		return nil
	}

	// Append the node to the existing list
	updatedNodes := append(upgrade.Status.UpgradedNodes, nodeName)

	// Create a patch with only the updated nodes
	patch := map[string]interface{}{
		"status": map[string]interface{}{
			"upgradedNodes": updatedNodes,
		},
	}

	updatedUpgrade, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal upgrade body: %w", err)
	}

	log.Info("Marking node as upgraded", "patch", patch)

	restClient, err := c.K8sdIoRestClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}
	result := restClient.Patch(types.MergePatchType).
		AbsPath("/apis/k8sd.io/v1alpha/upgrades/" + upgrade.Metadata.Name + "/status").
		Body(updatedUpgrade).
		Do(ctx)
	if result.Error() != nil {
		responseBody, _ := result.Raw()
		log.Error(result.Error(), "failed to update upgrade status", "response", string(responseBody))
		return fmt.Errorf("failed to update upgrade status: %w", result.Error())
	}

	return nil
}

// It returns an error if there is no upgrade in progress.
func (c *Client) SetUpgradePhase(ctx context.Context, phase string) error {
	log := log.FromContext(ctx).WithValues("upgrades", "SetUpgradePhase", "phase", phase)

	upgrade, err := c.GetInProgressUpgrade(ctx)
	if err != nil {
		return fmt.Errorf("failed to get in-progress upgrade: %w", err)
	}

	if upgrade == nil {
		return fmt.Errorf("no upgrade in progress")
	}

	upgrade.Status.Phase = phase
	updatedUpgrade, err := json.Marshal(upgrade)
	if err != nil {
		return fmt.Errorf("failed to marshal upgrade body: %w", err)
	}

	restClient, err := c.K8sdIoRestClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}
	result := restClient.Patch(types.MergePatchType).
		AbsPath("/apis/k8sd.io/v1alpha/upgrades/" + upgrade.Metadata.Name + "/status").
		Body(updatedUpgrade).
		Do(ctx)
	if result.Error() != nil {
		responseBody, _ := result.Raw()
		log.Error(result.Error(), "failed to update upgrade status phase", "response", string(responseBody))
		return fmt.Errorf("failed to update upgrade status phase: %w", result.Error())
	}

	return nil
}
