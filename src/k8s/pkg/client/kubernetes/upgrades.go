package kubernetes

import (
	"context"
	"fmt"
	"sort"

	"github.com/canonical/k8s/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO(Hue): move the upgrade CRD to a better location (maybe a separate package?)

// TODO: If the upgrade CRD grows, consider using kubebuilder.
const (
	UpgradePhaseNodeUpgrade    = "NodeUpgrade"
	UpgradePhaseFeatureUpgrade = "FeatureUpgrade"
	UpgradePhaseFailed         = "Failed"
	UpgradePhaseCompleted      = "Completed"
)

const (
	kind       = "Upgrade"
	group      = "k8sd.io"
	version    = "v1alpha"
	apiVersion = group + "/" + version
)

var (
	schemeGroupVersion = schema.GroupVersion{Group: group, Version: version}
)

type UpgradeStatus struct {
	Phase         string   `json:"phase,omitempty"`
	UpgradedNodes []string `json:"upgradedNodes,omitempty"`
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

// addUpgradeTypes registers upgrade types into the scheme.
func addUpgradeTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(schemeGroupVersion,
		&Upgrade{},
		&UpgradeList{},
	)
	metav1.AddToGroupVersion(scheme, schemeGroupVersion)
	return nil
}

func NewUpgrade(name string) Upgrade {
	return Upgrade{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiVersion,
			Kind:       kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: UpgradeStatus{Phase: UpgradePhaseNodeUpgrade, UpgradedNodes: []string{}},
	}
}

// GetInProgressUpgrade returns the upgrade CR that is currently in progress.
// TODO(ben): (KU-3218) Maybe make this more generic, e.g. GetUpgrade(filterFunc func(Upgrade) bool) (*Upgrade, error)
func (c *Client) GetInProgressUpgrade(ctx context.Context) (*Upgrade, error) {
	log := log.FromContext(ctx).WithValues("upgrades", "GetInProgressUpgrade")

	result := &UpgradeList{}
	if err := c.List(ctx, result); err != nil {
		if apierrors.IsNotFound(err) {
			// No upgrade in progress.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get upgrades: %w", err)
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
	if err := c.Create(ctx, &upgrade); err != nil {
		return fmt.Errorf("failed to create upgrade: %w", err)
	}

	// The status field needs to be patches separatly since it is a subresource.
	if err := c.PatchUpgradeStatus(ctx, upgrade.Name, upgrade.Status); err != nil {
		return fmt.Errorf("failed to patch upgrade status: %w", err)
	}

	return nil
}

// PatchUpgradeStatus patches the status of an upgrade CR.
func (c *Client) PatchUpgradeStatus(ctx context.Context, upgradeName string, status UpgradeStatus) error {
	u := NewUpgrade(upgradeName)
	p := ctrlclient.MergeFrom(&u)
	u.Status = status
	if err := c.Status().Patch(ctx, &u, p); err != nil {
		return fmt.Errorf("failed to patch upgrade status: %w", err)
	}

	return nil
}
