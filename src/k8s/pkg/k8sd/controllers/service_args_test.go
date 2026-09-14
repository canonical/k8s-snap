package controllers_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestServiceArgsController(t *testing.T) {
	newCtrl := func(s *mock.Snap, services []string, getRunningArgs func(context.Context, string) (map[string]string, error)) (*controllers.ServiceArgsController, chan time.Time) {
		triggerCh := make(chan time.Time)
		ctrl := controllers.NewServiceArgsController(controllers.ServiceArgsControllerOpts{
			Snap:           s,
			Services:       services,
			TriggerCh:      triggerCh,
			GetRunningArgs: getRunningArgs,
		})
		return ctrl, triggerCh
	}

	t.Run("NoRestartWhenArgsMatch", func(t *testing.T) {
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dir := t.TempDir()
		g.Expect(os.WriteFile(filepath.Join(dir, "kubelet"), []byte("--foo=\"bar\"\n"), 0o600)).To(Succeed())

		s := &mock.Snap{Mock: mock.Mock{ServiceArgumentsDir: dir}}
		ctrl, triggerCh := newCtrl(s, []string{"kubelet"}, func(_ context.Context, _ string) (map[string]string, error) {
			return map[string]string{"--foo": "bar"}, nil
		})
		go ctrl.Run(ctx)

		triggerCh <- time.Now()
		select {
		case <-ctrl.ReconciledCh():
		case <-time.After(channelSendTimeout):
			g.Fail("timed out waiting for reconciliation")
		}

		g.Expect(s.RestartServicesCalledWith).To(BeEmpty())
	})

	t.Run("RestartsWhenArgValueDiffers", func(t *testing.T) {
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dir := t.TempDir()
		g.Expect(os.WriteFile(filepath.Join(dir, "kubelet"), []byte("--foo=\"new-val\"\n"), 0o600)).To(Succeed())

		s := &mock.Snap{Mock: mock.Mock{ServiceArgumentsDir: dir}}
		ctrl, triggerCh := newCtrl(s, []string{"kubelet"}, func(_ context.Context, _ string) (map[string]string, error) {
			return map[string]string{"--foo": "old-val"}, nil
		})
		go ctrl.Run(ctx)

		triggerCh <- time.Now()
		select {
		case <-ctrl.ReconciledCh():
		case <-time.After(channelSendTimeout):
			g.Fail("timed out waiting for reconciliation")
		}

		g.Expect(s.RestartServicesCalledWith).To(HaveLen(1))
		g.Expect(s.RestartServicesCalledWith[0]).To(ConsistOf("kubelet"))
	})

	t.Run("RestartsWhenArgMissingFromRunningProcess", func(t *testing.T) {
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dir := t.TempDir()
		g.Expect(os.WriteFile(filepath.Join(dir, "kubelet"), []byte("--foo=\"bar\"\n"), 0o600)).To(Succeed())

		s := &mock.Snap{Mock: mock.Mock{ServiceArgumentsDir: dir}}
		ctrl, triggerCh := newCtrl(s, []string{"kubelet"}, func(_ context.Context, _ string) (map[string]string, error) {
			return map[string]string{}, nil // process has no args
		})
		go ctrl.Run(ctx)

		triggerCh <- time.Now()
		select {
		case <-ctrl.ReconciledCh():
		case <-time.After(channelSendTimeout):
			g.Fail("timed out waiting for reconciliation")
		}

		g.Expect(s.RestartServicesCalledWith).To(HaveLen(1))
		g.Expect(s.RestartServicesCalledWith[0]).To(ConsistOf("kubelet"))
	})

	t.Run("RestartsWhenRunningProcessHasExtraArg", func(t *testing.T) {
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dir := t.TempDir()
		// args file has only --foo; running process still has stale --removed arg
		g.Expect(os.WriteFile(filepath.Join(dir, "kubelet"), []byte("--foo=\"bar\"\n"), 0o600)).To(Succeed())

		s := &mock.Snap{Mock: mock.Mock{ServiceArgumentsDir: dir}}
		ctrl, triggerCh := newCtrl(s, []string{"kubelet"}, func(_ context.Context, _ string) (map[string]string, error) {
			return map[string]string{"--foo": "bar", "--removed": "old"}, nil
		})
		go ctrl.Run(ctx)

		triggerCh <- time.Now()
		select {
		case <-ctrl.ReconciledCh():
		case <-time.After(channelSendTimeout):
			g.Fail("timed out waiting for reconciliation")
		}

		g.Expect(s.RestartServicesCalledWith).To(HaveLen(1))
		g.Expect(s.RestartServicesCalledWith[0]).To(ConsistOf("kubelet"))
	})

	t.Run("NoRestartWhenProcessNotRunning", func(t *testing.T) {
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dir := t.TempDir()
		g.Expect(os.WriteFile(filepath.Join(dir, "kubelet"), []byte("--foo=\"bar\"\n"), 0o600)).To(Succeed())

		s := &mock.Snap{Mock: mock.Mock{ServiceArgumentsDir: dir}}
		ctrl, triggerCh := newCtrl(s, []string{"kubelet"}, func(_ context.Context, _ string) (map[string]string, error) {
			return nil, utils.ErrUnitNotRunning
		})
		go ctrl.Run(ctx)

		triggerCh <- time.Now()
		select {
		case <-ctrl.ReconciledCh():
		case <-time.After(channelSendTimeout):
			g.Fail("timed out waiting for reconciliation")
		}

		g.Expect(s.RestartServicesCalledWith).To(BeEmpty())
	})

	t.Run("NoRestartWhenArgsFileMissing", func(t *testing.T) {
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := &mock.Snap{Mock: mock.Mock{ServiceArgumentsDir: t.TempDir()}}
		ctrl, triggerCh := newCtrl(s, []string{"kubelet"}, func(_ context.Context, _ string) (map[string]string, error) {
			return map[string]string{"--foo": "bar"}, nil
		})
		go ctrl.Run(ctx)

		triggerCh <- time.Now()
		select {
		case <-ctrl.ReconciledCh():
		case <-time.After(channelSendTimeout):
			g.Fail("timed out waiting for reconciliation")
		}

		g.Expect(s.RestartServicesCalledWith).To(BeEmpty())
	})

	t.Run("ContinuesAfterGetRunningArgsError", func(t *testing.T) {
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dir := t.TempDir()
		for _, svc := range []string{"kubelet", "containerd"} {
			g.Expect(os.WriteFile(filepath.Join(dir, svc), []byte("--foo=\"bar\"\n"), 0o600)).To(Succeed())
		}

		s := &mock.Snap{Mock: mock.Mock{ServiceArgumentsDir: dir}}
		ctrl, triggerCh := newCtrl(s, []string{"kubelet", "containerd"}, func(_ context.Context, svc string) (map[string]string, error) {
			if svc == "kubelet" {
				return nil, fmt.Errorf("proc read error")
			}
			return map[string]string{"--foo": "old-val"}, nil // differs → should restart
		})
		go ctrl.Run(ctx)

		triggerCh <- time.Now()
		select {
		case <-ctrl.ReconciledCh():
		case <-time.After(channelSendTimeout):
			g.Fail("timed out waiting for reconciliation")
		}

		// kubelet error was swallowed, containerd should be restarted
		g.Expect(s.RestartServicesCalledWith).To(HaveLen(1))
		g.Expect(s.RestartServicesCalledWith[0]).To(ConsistOf("containerd"))
	})

	t.Run("OnlyDriftingServicesRestarted", func(t *testing.T) {
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dir := t.TempDir()
		for _, svc := range []string{"kubelet", "containerd"} {
			g.Expect(os.WriteFile(filepath.Join(dir, svc), []byte("--foo=\"bar\"\n"), 0o600)).To(Succeed())
		}

		s := &mock.Snap{Mock: mock.Mock{ServiceArgumentsDir: dir}}
		ctrl, triggerCh := newCtrl(s, []string{"kubelet", "containerd"}, func(_ context.Context, svc string) (map[string]string, error) {
			if svc == "kubelet" {
				return map[string]string{"--foo": "bar"}, nil // matches
			}
			return map[string]string{"--foo": "old-val"}, nil // drifted
		})
		go ctrl.Run(ctx)

		triggerCh <- time.Now()
		select {
		case <-ctrl.ReconciledCh():
		case <-time.After(channelSendTimeout):
			g.Fail("timed out waiting for reconciliation")
		}

		g.Expect(s.RestartServicesCalledWith).To(HaveLen(1))
		g.Expect(s.RestartServicesCalledWith[0]).To(ConsistOf("containerd"))
	})

	t.Run("StopsOnContextCancellation", func(t *testing.T) {
		g := NewWithT(t)
		ctx, cancel := context.WithCancel(context.Background())

		s := &mock.Snap{Mock: mock.Mock{ServiceArgumentsDir: t.TempDir()}}
		ctrl, _ := newCtrl(s, nil, nil)

		done := make(chan struct{})
		go func() {
			ctrl.Run(ctx)
			close(done)
		}()

		cancel()
		select {
		case <-done:
		case <-time.After(channelSendTimeout):
			g.Fail("controller did not stop after context cancellation")
		}
	})
}
