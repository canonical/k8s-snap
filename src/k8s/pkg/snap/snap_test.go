package snap_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestSnap(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewSnap(snap.SnapOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})

		err := snap.StartService(context.Background(), "test-service")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("snapctl start --enable k8s.test-service"))

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
		snap := snap.NewSnap(snap.SnapOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})
		err := snap.StopService(context.Background(), "test-service")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("snapctl stop --disable k8s.test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "test-service")
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Restart", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewSnap(snap.SnapOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})

		err := snap.RestartService(context.Background(), "test-service")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("snapctl restart k8s.test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "service")
			g.Expect(err).To(HaveOccurred())
		})
	})
}
