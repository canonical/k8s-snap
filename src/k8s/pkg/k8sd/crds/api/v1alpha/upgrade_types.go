// +kubebuilder:validation:Required

package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=NodeUpgrade;FeatureUpgrade;Completed;Failed
// +kubebuilder:validation:MinLength=1
type UpgradePhase string

// +kubebuilder:validation:Enum=RollingUpgrade;RollingDowngrade;InPlace
// +kubebuilder:validation:MinLength=1
type UpgradeStrategy string

// NOTE(Hue): Make sure to keep these up to date with the UpgradePhase type
// and UpgradeStrategy type Enum validations.
const (
	UpgradePhaseNodeUpgrade    UpgradePhase = "NodeUpgrade"
	UpgradePhaseFeatureUpgrade UpgradePhase = "FeatureUpgrade"
	UpgradePhaseCompleted      UpgradePhase = "Completed"
	UpgradePhaseFailed         UpgradePhase = "Failed"

	UpgradeStrategyRollingUpgrade   UpgradeStrategy = "RollingUpgrade"
	UpgradeStrategyRollingDowngrade UpgradeStrategy = "RollingDowngrade"
	UpgradeStrategyInPlace          UpgradeStrategy = "InPlace"
)

// UpgradeStatus defines the observed state of Upgrade.
type UpgradeStatus struct {
	// Phase indicates the current phase of the upgrade process.
	Phase UpgradePhase `json:"phase,omitempty"`
	// Strategy indicates the strategy used for the upgrade.
	Strategy UpgradeStrategy `json:"strategy,omitempty"`
	// UpgradedNodes is a list of nodes that have been successfully upgraded.
	// +optional
	UpgradedNodes []string `json:"upgradedNodes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Strategy",type="string",JSONPath=".status.strategy"

// Upgrade is the Schema for the upgrades API.
type Upgrade struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status UpgradeStatus `json:"status,omitempty"`
}

// NewUpgrade creates a new Upgrade object with the given name.
func NewUpgrade(name string) *Upgrade {
	return &Upgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// +kubebuilder:object:root=true

// UpgradeList contains a list of Upgrade.
type UpgradeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Upgrade `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Upgrade{}, &UpgradeList{})
}
