package kubernetes

import (
	"context"
	"fmt"
	"sort"

	upgradesv1alpha "github.com/canonical/k8s/pkg/k8sd/crds/upgrades/v1alpha"
	"github.com/canonical/k8s/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO(ben): (KU-3218) Maybe make this more generic, e.g. GetUpgrade(filterFunc func(Upgrade) bool) (*Upgrade, error).
// GetInProgressUpgrade returns the upgrade CR that is currently in progress.
func (c *Client) GetInProgressUpgrade(ctx context.Context) (*upgradesv1alpha.Upgrade, error) {
	log := log.FromContext(ctx).WithValues("upgrades", "GetInProgressUpgrade")

	result := &upgradesv1alpha.UpgradeList{}
	if err := c.List(ctx, result); err != nil {
		if apierrors.IsNotFound(err) {
			// No upgrade in progress.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get upgrades: %w", err)
	}

	var matches []upgradesv1alpha.Upgrade
	for _, upgrade := range result.Items {
		if upgrade.Status.Phase != upgradesv1alpha.UpgradePhaseFailed && upgrade.Status.Phase != upgradesv1alpha.UpgradePhaseCompleted {
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

// PatchUpgradeStatus patches the status of an upgrade CR.
func (c *Client) PatchUpgradeStatus(ctx context.Context, u *upgradesv1alpha.Upgrade, status upgradesv1alpha.UpgradeStatus) error {
	p := ctrlclient.MergeFrom(u.DeepCopy())
	u.Status = status
	if err := c.Status().Patch(ctx, u, p); err != nil {
		return fmt.Errorf("failed to patch upgrade status: %w", err)
	}

	return nil
}
