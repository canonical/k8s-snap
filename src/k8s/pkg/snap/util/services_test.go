package snaputil

import (
	"context"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestStartWorkerServices(t *testing.T) {
	mock := &mock.Snap{
		Mock: mock.Mock{},
	}
	g := NewGomegaWithT(t)

	mock.StartServiceErr = fmt.Errorf("service start failed")

	t.Run("AllServicesStartSuccess", func(t *testing.T) {
		mock.StartServiceErr = nil
		g.Expect(StartWorkerServices(context.Background(), mock)).To(Succeed())
		g.Expect(mock.StartServiceCalledWith).To(ConsistOf(workerServices))
	})

	t.Run("ServiceStartFailure", func(t *testing.T) {
		mock.StartServiceErr = fmt.Errorf("service start failed")
		g.Expect(StartWorkerServices(context.Background(), mock)).NotTo(Succeed())
	})
}

func TestStartControlPlaneServices(t *testing.T) {
	mock := &mock.Snap{
		Mock: mock.Mock{},
	}
	g := NewGomegaWithT(t)

	mock.StartServiceErr = fmt.Errorf("service start failed")

	t.Run("AllServicesStartSuccess", func(t *testing.T) {
		mock.StartServiceErr = nil
		g.Expect(StartControlPlaneServices(context.Background(), mock)).To(Succeed())
		g.Expect(mock.StartServiceCalledWith).To(ConsistOf(controlPlaneServices))
	})

	t.Run("ServiceStartFailure", func(t *testing.T) {
		mock.StartServiceErr = fmt.Errorf("service start failed")
		g.Expect(StartControlPlaneServices(context.Background(), mock)).NotTo(Succeed())
	})
}

func TestStartK8sDqliteServices(t *testing.T) {
	mock := &mock.Snap{
		Mock: mock.Mock{},
	}
	g := NewGomegaWithT(t)

	mock.StartServiceErr = fmt.Errorf("service start failed")

	t.Run("ServiceStartSuccess", func(t *testing.T) {
		mock.StartServiceErr = nil
		g.Expect(StartK8sDqliteServices(context.Background(), mock)).To(Succeed())
		g.Expect(mock.StartServiceCalledWith).To(ConsistOf("k8s-dqlite"))
	})

	t.Run("ServiceStartFailure", func(t *testing.T) {
		mock.StartServiceErr = fmt.Errorf("service start failed")
		g.Expect(StartK8sDqliteServices(context.Background(), mock)).NotTo(Succeed())
	})
}

func TestStopControlPlaneServices(t *testing.T) {
	mock := &mock.Snap{
		Mock: mock.Mock{},
	}
	g := NewGomegaWithT(t)

	mock.StopServiceErr = fmt.Errorf("service stop failed")

	t.Run("AllServicesStopSuccess", func(t *testing.T) {
		mock.StopServiceErr = nil
		g.Expect(StopControlPlaneServices(context.Background(), mock)).To(Succeed())
		g.Expect(mock.StopServiceCalledWith).To(ConsistOf(controlPlaneServices))
	})

	t.Run("ServiceStopFailure", func(t *testing.T) {
		mock.StopServiceErr = fmt.Errorf("service stop failed")
		g.Expect(StopControlPlaneServices(context.Background(), mock)).NotTo(Succeed())
	})
}

func TestStopK8sDqliteServices(t *testing.T) {
	mock := &mock.Snap{
		Mock: mock.Mock{},
	}
	g := NewGomegaWithT(t)

	mock.StopServiceErr = fmt.Errorf("service stop failed")

	t.Run("ServiceStopSuccess", func(t *testing.T) {
		mock.StopServiceErr = nil
		g.Expect(StopK8sDqliteServices(context.Background(), mock)).To(Succeed())
		g.Expect(mock.StopServiceCalledWith).To(ConsistOf("k8s-dqlite"))
	})

	t.Run("ServiceStopFailure", func(t *testing.T) {
		mock.StopServiceErr = fmt.Errorf("service stop failed")
		g.Expect(StopK8sDqliteServices(context.Background(), mock)).NotTo(Succeed())
	})
}
