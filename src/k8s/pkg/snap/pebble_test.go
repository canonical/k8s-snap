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

		err := snap.StartServices(context.Background(), []string{"test-service"})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble start test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartServices(context.Background(), []string{"test-service"})
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
		err := snap.StopServices(context.Background(), []string{"test-service"})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble stop test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartServices(context.Background(), []string{"test-service"})
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

		err := snap.RestartServices(context.Background(), []string{"test-service"})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble restart test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartServices(context.Background(), []string{"service"})
			g.Expect(err).To(HaveOccurred())
		})
	})
}
