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

func (in *Upgrade) DeepCopyInto(out *Upgrade) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Status.Phase = in.Status.Phase
	out.Status.UpgradedNodes = make([]string, len(in.Status.UpgradedNodes))
	copy(out.Status.UpgradedNodes, in.Status.UpgradedNodes)
}

func (in *Upgrade) DeepCopy() *Upgrade {
	if in == nil {
		return nil
	}
	out := new(Upgrade)
	in.DeepCopyInto(out)
	return out
}

func (in *Upgrade) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// UpgradeList contains a list of Upgrade resources.
type UpgradeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Upgrade `json:"items"`
}

func (in *UpgradeList) DeepCopyInto(out *UpgradeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Upgrade, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *UpgradeList) DeepCopy() *UpgradeList {
	if in == nil {
		return nil
	}
	out := new(UpgradeList)
	in.DeepCopyInto(out)
	return out
}

func (in *UpgradeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
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

	return nil
}

// PatchUpgradeStatus patches the status of an upgrade CR.
func (c *Client) PatchUpgradeStatus(ctx context.Context, u *Upgrade, status UpgradeStatus) error {
	p := ctrlclient.MergeFrom(u.DeepCopy())
	u.Status = status
	log.FromContext(ctx).Info("patching upgrade status", "phase", u.Status.Phase, "upgradedNodes", u.Status.UpgradedNodes)
	if err := c.Status().Patch(ctx, u, p); err != nil {
		return fmt.Errorf("failed to patch upgrade status: %w", err)
	}

	return nil
}
