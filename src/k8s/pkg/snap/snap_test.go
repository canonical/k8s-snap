package snap

import (
	"context"
	"fmt"
	"testing"

	. "github.com/canonical/k8s/pkg/snap/mock/runner"

	. "github.com/onsi/gomega"
)

func TestServiceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"With k8s. prefix", "k8s.test-service", "k8s.test-service"},
		{"Without prefix", "api", "k8s.api"},
		{"Just k8s", "k8s", "k8s"},
		{"Empty string", "", "k8s."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := serviceName(tc.input)
			if got != tc.expected {
				t.Errorf("serviceName(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestServices(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &MockRunner{}
		snap := NewSnap("testdir", "testdir", WithCommandRunner(mockRunner.Run))

		err := snap.StartService(context.Background(), "test-service")
		g.Expect(err).To(BeNil())
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("snapctl start --enable k8s.test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "test-service")
			g.Expect(err).NotTo(BeNil())
		})
	})

	t.Run("Stop", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &MockRunner{}
		snap := NewSnap("testdir", "testdir", WithCommandRunner(mockRunner.Run))

		err := snap.StopService(context.Background(), "test-service")
		g.Expect(err).To(BeNil())
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("snapctl stop --disable k8s.test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "test-service")
			g.Expect(err).NotTo(BeNil())
		})
	})

	t.Run("Restart", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &MockRunner{}
		snap := NewSnap("testdir", "testdir", WithCommandRunner(mockRunner.Run))

		err := snap.RestartService(context.Background(), "test-service")
		g.Expect(err).To(BeNil())
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("snapctl restart k8s.test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "service")
			g.Expect(err).NotTo(BeNil())
		})
	})
}
