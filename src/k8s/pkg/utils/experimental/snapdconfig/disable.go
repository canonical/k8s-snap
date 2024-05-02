package snapdconfig

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
)

func Disable(ctx context.Context, s snap.Snap) error {
	b, err := json.Marshal(Meta{Orb: "none", APIVersion: "1.30"})
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := s.SnapctlSet(ctx, fmt.Sprintf("meta=%s", string(b)), "dns!", "network!", "gateway!", "ingress!", "load-balancer!", "local-storage!"); err != nil {
		return fmt.Errorf("failed to snapctl set: %w", err)
	}
	return nil
}
