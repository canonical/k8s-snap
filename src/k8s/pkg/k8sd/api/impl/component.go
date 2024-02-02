package impl

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/component"
	"github.com/canonical/k8s/pkg/snap"
)

// GetComponent returns the current status of the k8s components.
func GetComponents(snap snap.Snap) ([]api.Component, error) {
	manager, err := component.NewHelmClient(snap)
	if err != nil {
		return nil, fmt.Errorf("failed to get component manager: %w", err)
	}

	components, err := manager.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}

	// Decouple the internal Component type from the external API types.
	extComponents := make([]api.Component, len(components))
	for i, component := range components {
		var status api.ComponentStatus
		if component.Status {
			status = api.ComponentEnable
		} else {
			status = api.ComponentDisable
		}

		extComponents[i] = api.Component{
			Name:   component.Name,
			Status: status,
		}
	}
	return extComponents, nil
}
