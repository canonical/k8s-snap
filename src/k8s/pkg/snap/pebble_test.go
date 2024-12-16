package snap_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestPebble(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewPebble(snap.PebbleOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})

		err := snap.StartService(context.Background(), "test-service")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble start test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "test-service")
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Stop", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewPebble(snap.PebbleOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})
		err := snap.StopService(context.Background(), "test-service")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble stop test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "test-service")
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("GetServiceState", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{
			RunOutput: "Service  Startup  Current   Since\ntest-service  enabled  active  -",
		}
		snap := snap.NewPebble(snap.PebbleOpts{
			SnapDir:    "testdir",
			RunCommand: mockRunner.Run,
		})

		state, err := snap.GetServiceState(context.Background(), "test-service")

		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(state).To(Equal("active"))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble services test-service"))

		t.Run("run error", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			_, err := snap.GetServiceState(context.Background(), "test-service")

			g.Expect(err).To(HaveOccurred())
		})

		t.Run("invalid output", func(t *testing.T) {
			g := NewWithT(t)

			mockRunner.RunOutput = "Service  Startup  Current   Since"
			_, err := snap.GetServiceState(context.Background(), "test-service")
			g.Expect(err).To(HaveOccurred())

			mockRunner.RunOutput = "Service  Startup  Current   Since\nFoo"
			_, err = snap.GetServiceState(context.Background(), "test-service")
			g.Expect(err).To(HaveOccurred())

			mockRunner.RunOutput = "Service  Startup  Current   Since\ntest-service enabled foo -"
			_, err = snap.GetServiceState(context.Background(), "test-service")
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Restart", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewPebble(snap.PebbleOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})

		err := snap.RestartService(context.Background(), "test-service")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble restart test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "service")
			g.Expect(err).To(HaveOccurred())
		})
	})
}
