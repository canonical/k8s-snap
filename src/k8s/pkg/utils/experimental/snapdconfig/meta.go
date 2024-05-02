package snapdconfig

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
)

// Meta represents meta configuration that describes how to parse the snapd configuration values.
// See docs/snapd-config.md for details.
type Meta struct {
	// Orb is one of "", "k8sd", "snapd", "none".
	Orb string `json:"orb"`
	// APIVersion is one of "", "1.30".
	APIVersion string `json:"apiVersion"`
	// Error is an error message that is set if the config mode cannot be parsed.
	Error string `json:"error,omitempty"`
}

// ParseMeta parses the output of "snapctl get -d meta" and returns the active snapd configuration mode.
// ParseMeta returns an error and true if the error is because of missing/empty config, rather than an operational error.
func ParseMeta(ctx context.Context, s snap.Snap) (Meta, bool, error) {
	var parse struct {
		Meta Meta `json:"meta"`
	}

	b, err := s.SnapctlGet(ctx, "-d", "meta")
	if err != nil {
		return Meta{}, false, fmt.Errorf("failed to get snapd config mode: %w", err)
	}
	if err := json.Unmarshal(b, &parse); err != nil {
		return Meta{}, false, fmt.Errorf("failed to parse snap config mode: %w", err)
	}

	// default meta.orb is none
	if parse.Meta.Orb == "" {
		parse.Meta.Orb = "none"
	}
	switch parse.Meta.Orb {
	case "k8sd", "snapd", "none":
	default:
		return Meta{}, false, fmt.Errorf("invalid meta.orb value %q", parse.Meta.Orb)
	}

	// default meta.version is 1.30
	if parse.Meta.APIVersion == "" {
		parse.Meta.APIVersion = "1.30"
	}
	switch parse.Meta.APIVersion {
	case "1.30":
	default:
		return Meta{}, false, fmt.Errorf("invalid meta.apiVersion value %q", parse.Meta.APIVersion)
	}

	return parse.Meta, true, nil
}

// SetMeta sets the active snapd configuration mode.
func SetMeta(ctx context.Context, s snap.Snap, meta Meta) error {
	b, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := s.SnapctlSet(ctx, fmt.Sprintf("meta=%s", string(b))); err != nil {
		return fmt.Errorf("failed to snapctl set meta: %w", err)
	}
	return nil
}
