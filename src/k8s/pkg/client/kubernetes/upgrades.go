package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/canonical/k8s/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// TODO: If the upgrade CRD grows, consider using kubebuilder.type Metadata struct {
const (
	UpgradePhaseNodeUpgrade    = "NodeUpgrade"
	UpgradePhaseFeatureUpgrade = "FeatureUpgrade"
	UpgradePhaseFailed         = "Failed"
	UpgradePhaseCompleted      = "Completed"
)

const (
	kind    = "Upgrade"
	group   = "k8sd.io"
	version = "v1alpha"
)

type Metadata struct {
	Name string `json:"name,omitempty"`
}

type Status struct {
	Phase         string   `json:"phase,omitempty"`
	UpgradedNodes []string `json:"upgradedNodes,omitempty"`
}

type Upgrade struct {
	APIVersion string   `json:"apiVersion,omitempty"`
	Kind       string   `json:"kind,omitempty"`
	Metadata   Metadata `json:"metadata,omitempty"`
	Status     Status   `json:"status,omitempty"`
}

func NewUpgrade(name string) Upgrade {
	return Upgrade{
		APIVersion: group + "/" + version,
		Kind:       kind,
		Metadata:   Metadata{Name: name},
		Status:     Status{Phase: UpgradePhaseNodeUpgrade, UpgradedNodes: []string{}},
	}
}

func (c *Client) K8sdIoRestClient() (*rest.RESTClient, error) {
	k8sdConfig := c.RESTConfig()
	k8sdConfig.GroupVersion = &schema.GroupVersion{Group: group, Version: version}
	k8sdConfig.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())

	restClient, err := rest.RESTClientFor(k8sdConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}
	return restClient, nil
}

// GetInProgressUpgrade returns the upgrade CR that is currently in progress.
// TODO(ben): Maybe make this more generic, e.g. GetUpgrade(filterFunc func(Upgrade) bool) (*Upgrade, error)
func (c *Client) GetInProgressUpgrade(ctx context.Context) (*Upgrade, error) {
	log := log.FromContext(ctx).WithValues("upgrades", "GetInProgressUpgrade")

	restClient, err := c.K8sdIoRestClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}

	upgrades, err := restClient.Get().AbsPath(fmt.Sprintf("/apis/%s/%s/upgrades", group, version)).DoRaw(ctx)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// No upgrade in progress.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get upgrades: %w", err)
	}

	log.Info("Got upgrades", "upgrades", string(upgrades))
	var result struct {
		Items []Upgrade `json:"items"`
	}
	if err := json.Unmarshal(upgrades, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upgrades: %w", err)
	}

	var matches []Upgrade
	for _, upgrade := range result.Items {
		if upgrade.Status.Phase != UpgradePhaseFailed && upgrade.Status.Phase != UpgradePhaseCompleted {
			matches = append(matches, upgrade)
		}
	}
	if len(matches) == 0 {
		return nil, nil
	}
	if len(matches) > 1 {
		log.Info("Warning: Found multiple in-progress upgrades", "inprogress upgrades", len(matches))
	}
	// Sort matches by name
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Metadata.Name < matches[j].Metadata.Name
	})

	// Return the latest
	return &matches[len(matches)-1], nil
}

// CreateUpgrade creates a new upgrade CR.
func (c *Client) CreateUpgrade(ctx context.Context, upgrade Upgrade) error {
	log := log.FromContext(ctx).WithValues("upgrades", "createUpgrade")
	restClient, err := c.K8sdIoRestClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}

	body, err := json.Marshal(upgrade)
	if err != nil {
		return fmt.Errorf("failed to marshal upgrade: %w", err)
	}

	log.Info("Creating upgrade", "upgrade", upgrade)
	result := restClient.Post().
		AbsPath(fmt.Sprintf("/apis/%s/%s/upgrades", group, version)).
		Body(body).
		Do(ctx)
	if result.Error() != nil {
		responseBody, _ := result.Raw()
		log.Error(result.Error(), "failed to create upgrade", "response", string(responseBody))
		return fmt.Errorf("failed to create upgrade: %w", result.Error())
	}

	// The status field needs to be patches separatly since it is a subresource.
	if err := c.PatchUpgradeStatus(ctx, upgrade.Metadata.Name, upgrade.Status); err != nil {
		return fmt.Errorf("failed to patch upgrade status: %w", err)
	}

	return nil
}

// PatchUpgradeStatus patches the status of an upgrade CR.
func (c *Client) PatchUpgradeStatus(ctx context.Context, upgradeName string, status Status) error {
	log := log.FromContext(ctx).WithValues("upgrades", "PatchUpgrade", "upgrade", upgradeName, "status", status)

	restClient, err := c.K8sdIoRestClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}

	// Wrap the status in a struct to match the CRD definition.
	upgrade := NewUpgrade(upgradeName)
	upgrade.Status = status

	body, err := json.Marshal(upgrade)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	log.WithValues("upgrade", upgrade).Info("Patching upgrade")
	result := restClient.Patch(types.MergePatchType).
		AbsPath(fmt.Sprintf("/apis/%s/%s/upgrades/%s/status", group, version, upgradeName)).
		Body(body).
		Do(ctx)
	if result.Error() != nil {
		responseBody, _ := result.Raw()
		log.Error(result.Error(), "failed to update upgrade status", "response", string(responseBody))
		return fmt.Errorf("failed to update upgrade status: %w", result.Error())
	}

	return nil
}
