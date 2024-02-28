package snap_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/snap/mock"

	. "github.com/onsi/gomega"
)

func TestServices(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewSnap("testdir", "testdir", snap.WithCommandRunner(mockRunner.Run))

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
		mockRunner := &mock.Runner{}
		snap := snap.NewSnap("testdir", "testdir", snap.WithCommandRunner(mockRunner.Run))

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
		mockRunner := &mock.Runner{}
		snap := snap.NewSnap("testdir", "testdir", snap.WithCommandRunner(mockRunner.Run))

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
