package setup_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	. "github.com/onsi/gomega"
)

func TestApplyStickyBitsToWorldWritableDirectories(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	// This test just ensures the function doesn't panic or error unexpectedly.
	// The actual chmod operation requires root privileges and is difficult to test in unit tests.
	// The function is designed to be non-fatal even if it fails.
	err := setup.ApplyStickyBitsToWorldWritableDirectories(ctx)
	if err != nil {
		t.Logf("ApplyStickyBitsToWorldWritableDirectories returned an error (this may be expected in non-root tests): %v", err)
	}

	cmd := exec.CommandContext(ctx, "bash", "-c",
		`df --local -P | awk '{if (NR!=1) print $6}' | xargs -I '$6' find '$6' -xdev -type d \( -perm -0002 -a ! -perm -1000 \)`)
	output, err := cmd.CombinedOutput()
	g.Expect(err).To(BeNil(), "Failed to find world-writable directories without sticky bits")
	g.Expect(string(output)).To(BeEmpty())

}
