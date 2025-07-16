package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/canonical/k8s/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// TODO: If the upgrade CRD grows, consider using kubebuilder.
type UpgradePhase string

const (
	UpgradePhaseNodeUpgrade    = "NodeUpgrade"
	UpgradePhaseFeatureUpgrade = "FeatureUpgrade"
	UpgradePhaseFailed         = "Failed"
	UpgradePhaseCompleted      = "Completed"
)

type UpgradeStrategy string

const (
	UpgradeStrategyRollingUpgrade   = "RollingUpgrade"
	UpgradeStrategyRollingDowngrade = "RollingDowngrade"
	UpgradeStrategyInPlace          = "InPlace"
)

const (
	kind            = "Upgrade"
	group           = "k8sd.io"
	version         = "v1alpha"
	apiVersion      = group + "/" + version
	upgradesAPIPath = "/apis/" + group + "/" + version + "/upgrades"
)

var (
	schemeGroupVersion = schema.GroupVersion{Group: group, Version: version}
	schemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme        = schemeBuilder.AddToScheme
)

type UpgradeStatus struct {
	Strategy      UpgradeStrategy `json:"strategy,omitempty"`
	Phase         UpgradePhase    `json:"phase,omitempty"`
	UpgradedNodes []string        `json:"upgradedNodes,omitempty"`
}

// TODO(Hue): (KU-3033) Use kubebuilder to generate the CRD .
type Upgrade struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status UpgradeStatus `json:"status,omitempty"`
}

func (u *Upgrade) DeepCopyObject() runtime.Object {
	if u == nil {
		return nil
	}
	cp := *u
	cp.ObjectMeta = *u.ObjectMeta.DeepCopy()
	cp.Status.Phase = u.Status.Phase
	if u.Status.UpgradedNodes != nil {
		nodesCopy := make([]string, len(u.Status.UpgradedNodes))
		copy(nodesCopy, u.Status.UpgradedNodes)
		cp.Status.UpgradedNodes = nodesCopy
	}
	return &cp
}

// UpgradeList contains a list of Upgrade resources.
type UpgradeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Upgrade `json:"items"`
}

func (ul *UpgradeList) DeepCopyObject() runtime.Object {
	if ul == nil {
		return nil
	}
	cp := *ul
	cp.ListMeta = *ul.ListMeta.DeepCopy()
	cp.Items = make([]Upgrade, len(ul.Items))
	for i, u := range ul.Items {
		uCp, ok := u.DeepCopyObject().(*Upgrade)
		if !ok {
			log.L().Error(fmt.Errorf("type assertion failed for upgrade deepcopy"), "upgrade", u)
			continue
		}
		cp.Items[i] = *uCp
	}
	return &cp
}

// addKnownTypes registers upgrade types into the scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(schemeGroupVersion,
		&Upgrade{},
		&UpgradeList{},
	)
	metav1.AddToGroupVersion(scheme, schemeGroupVersion)
	return nil
}

func NewUpgrade(name string, strategy UpgradeStrategy) Upgrade {
	return Upgrade{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: UpgradeStatus{
			Strategy:      strategy,
			Phase:         UpgradePhaseNodeUpgrade,
			UpgradedNodes: []string{},
		},
	}
}

func (c *Client) k8sdIoRestClient() (*rest.RESTClient, error) {
	k8sdConfig := c.RESTConfig()
	k8sdConfig.GroupVersion = &schemeGroupVersion
	k8sdConfig.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())

	restClient, err := rest.RESTClientFor(k8sdConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}
	return restClient, nil
}

// TODO(ben): (KU-3218) Maybe make this more generic, e.g. GetUpgrade(filterFunc func(Upgrade) bool) (*Upgrade, error).
// GetInProgressUpgrade returns the upgrade CR that is currently in progress.
func (c *Client) GetInProgressUpgrade(ctx context.Context) (*Upgrade, error) {
	log := log.FromContext(ctx).WithValues("upgrades", "GetInProgressUpgrade")

	restClient, err := c.k8sdIoRestClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}

	upgrades, err := restClient.Get().AbsPath(upgradesAPIPath).DoRaw(ctx)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// No upgrade in progress.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get upgrades: %w", err)
	}

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
	lenMatches := len(matches)
	if lenMatches == 0 {
		return nil, nil
	}
	if lenMatches > 1 {
		log.Info("Warning: Found multiple in-progress upgrades", "inprogress upgrades", len(matches))
	}
	// Sort matches by name
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})

	// Return the latest
	return &matches[lenMatches-1], nil
}

// CreateUpgrade creates a new upgrade CR.
func (c *Client) CreateUpgrade(ctx context.Context, upgrade Upgrade) error {
	log := log.FromContext(ctx).WithValues("upgrades", "createUpgrade")
	restClient, err := c.k8sdIoRestClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}

	body, err := json.Marshal(upgrade)
	if err != nil {
		return fmt.Errorf("failed to marshal upgrade: %w", err)
	}

	log.Info("Creating upgrade", "upgrade", upgrade)
	result := restClient.Post().
		AbsPath(upgradesAPIPath).
		Body(body).
		Do(ctx)
	if result.Error() != nil {
		responseBody, _ := result.Raw()
		log.Error(result.Error(), "failed to create upgrade", "response", string(responseBody))
		return fmt.Errorf("failed to create upgrade: %w", result.Error())
	}

	// The status field needs to be patches separatly since it is a subresource.
	if err := c.PatchUpgradeStatus(ctx, upgrade.Name, upgrade.Status); err != nil {
		return fmt.Errorf("failed to patch upgrade status: %w", err)
	}

	return nil
}

// PatchUpgradeStatus patches the status of an upgrade CR.
func (c *Client) PatchUpgradeStatus(ctx context.Context, upgradeName string, status UpgradeStatus) error {
	log := log.FromContext(ctx).WithValues("upgrades", "PatchUpgrade", "upgrade", upgradeName, "status", status)

	restClient, err := c.k8sdIoRestClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client for k8sd.io group: %w", err)
	}

	// Wrap the status in a struct to match the CRD definition.
	upgrade := NewUpgrade(upgradeName, status.Strategy)
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
