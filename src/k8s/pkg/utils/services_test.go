package utils

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	. "github.com/onsi/gomega"
)

func skipIfNoSystemctl(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Skip("systemctl not available")
	}
}

func TestServiceArgsFromMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]*string
		expected struct {
			updateArgs map[string]string
			deleteArgs []string
		}
	}{
		{
			name:  "NilValue",
			input: map[string]*string{"arg1": nil},
			expected: struct {
				updateArgs map[string]string
				deleteArgs []string
			}{
				updateArgs: map[string]string{},
				deleteArgs: []string{"arg1"},
			},
		},
		{
			name:  "EmptyString", // Should be threated as normal string
			input: map[string]*string{"arg1": Pointer("")},
			expected: struct {
				updateArgs map[string]string
				deleteArgs []string
			}{
				updateArgs: map[string]string{"arg1": ""},
				deleteArgs: []string{},
			},
		},
		{
			name:  "NonEmptyString",
			input: map[string]*string{"arg1": Pointer("value1")},
			expected: struct {
				updateArgs map[string]string
				deleteArgs []string
			}{
				updateArgs: map[string]string{"arg1": "value1"},
				deleteArgs: []string{},
			},
		},
		{
			name: "MixedValues",
			input: map[string]*string{
				"arg1": Pointer("value1"),
				"arg2": Pointer(""),
				"arg3": nil,
			},
			expected: struct {
				updateArgs map[string]string
				deleteArgs []string
			}{
				updateArgs: map[string]string{"arg1": "value1", "arg2": ""},
				deleteArgs: []string{"arg3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			updateArgs, deleteArgs := ServiceArgsFromMap(tt.input)
			g.Expect(updateArgs).To(Equal(tt.expected.updateArgs))
			g.Expect(deleteArgs).To(Equal(tt.expected.deleteArgs))
		})
	}
}

func TestGetUnitLoadState(t *testing.T) {
	skipIfNoSystemctl(t)
	ctx := context.Background()

	t.Run("NotFound", func(t *testing.T) {
		g := NewWithT(t)
		state, err := getUnitLoadState(ctx, "snap.k8s.definitely-nonexistent-k8sd-test.service")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(state).To(Equal(LoadState(LoadStateNotFound)))
	})

	t.Run("Loaded", func(t *testing.T) {
		// init.scope is always loaded and active on any systemd system.
		g := NewWithT(t)
		state, err := getUnitLoadState(ctx, "init.scope")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(state).To(Equal(LoadState(LoadStateLoaded)))
	})
}

func TestGetUnitActiveState(t *testing.T) {
	skipIfNoSystemctl(t)
	ctx := context.Background()

	t.Run("Active", func(t *testing.T) {
		// init.scope is always loaded and active on any systemd system.
		g := NewWithT(t)
		state, err := getUnitActiveState(ctx, "init.scope")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(state).To(Equal(ActiveStateActive))
	})

	t.Run("Inactive", func(t *testing.T) {
		// A loaded-but-inactive unit: systemd ships plymouth-quit.service which
		// is loaded but finishes quickly and stays inactive afterward.
		// Use a unit that is guaranteed to be loaded but not running:
		// systemd-ask-password-console.service is loaded but inactive on headless systems.
		// Fall back to checking any not-found unit — systemctl returns "inactive" for those too.
		g := NewWithT(t)
		state, err := getUnitActiveState(ctx, "snap.k8s.definitely-nonexistent-k8sd-test.service")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(state).To(Equal(ActiveStateInactive))
	})
}

func TestGetUnitMainPID(t *testing.T) {
	skipIfNoSystemctl(t)
	ctx := context.Background()

	t.Run("ActiveUnitHasNonZeroPID", func(t *testing.T) {
		// systemd-journald is always running on any systemd system.
		g := NewWithT(t)
		pid, err := getUnitMainPID(ctx, "systemd-journald.service")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(pid).To(BeNumerically(">", 0))
	})

	t.Run("InactiveUnitHasZeroPID", func(t *testing.T) {
		g := NewWithT(t)
		pid, err := getUnitMainPID(ctx, "snap.k8s.definitely-nonexistent-k8sd-test.service")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(pid).To(Equal(0))
	})
}

func TestRunningServiceArgs(t *testing.T) {
	skipIfNoSystemctl(t)
	ctx := context.Background()

	t.Run("UnitNotFound", func(t *testing.T) {
		// A completely unknown service name: LoadState=not-found.
		// This is NOT wrapped in ErrUnitNotRunning — it's a different error class.
		g := NewWithT(t)
		_, err := RunningServiceArgs(ctx, "definitely-nonexistent-k8sd-test-unit")
		g.Expect(err).To(HaveOccurred())
		g.Expect(errors.Is(err, ErrUnitNotRunning)).To(BeFalse())
		g.Expect(err.Error()).To(ContainSubstring("was not found"))
	})
}
