package internal_test

import (
	"context"
	"testing"

	"github.com/canonical/k8s/pkg/snap/util/cleanup/internal"
)

func TestLogoutISCSISessions_NoIscsiadm(t *testing.T) {
	// When iscsiadm is not available the function must return without panicking.
	// We cannot control PATH in a unit test environment, so we just verify it doesn't panic.
	t.Setenv("PATH", t.TempDir())
	internal.LogoutISCSISessions(context.Background())
}

func TestSyncISCSIDevices_NoIscsiadm(t *testing.T) {
	// When iscsiadm is not available the function must return without panicking.
	t.Setenv("PATH", t.TempDir())
	internal.SyncISCSIDevices(context.Background())
}
