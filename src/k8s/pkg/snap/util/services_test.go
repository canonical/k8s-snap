package snaputil

import (
	"context"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestStartWorkerServices(t *testing.T) {
	mock := &mock.Snap{
		Mock: mock.Mock{},
	}
	g := NewWithT(t)

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
	g := NewWithT(t)

	mock.StartServiceErr = fmt.Errorf("service start failed")

	t.Run("AllServicesStartSuccess", func(t *testing.T) {
		mock.StartServiceErr = nil
		g.Expect(StartControlPlaneServices(context.Background(), mock)).To(Succeed())
		g.Expect(mock.StartServiceCalledWith).To(ConsistOf(ControlPlaneServices))
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
	g := NewWithT(t)

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
	g := NewWithT(t)

	mock.StopServiceErr = fmt.Errorf("service stop failed")

	t.Run("AllServicesStopSuccess", func(t *testing.T) {
		mock.StopServiceErr = nil
		g.Expect(StopControlPlaneServices(context.Background(), mock)).To(Succeed())
		g.Expect(mock.StopServiceCalledWith).To(ConsistOf(ControlPlaneServices))
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
	g := NewWithT(t)

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
			input: map[string]*string{"arg1": utils.Pointer("")},
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
			input: map[string]*string{"arg1": utils.Pointer("value1")},
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
				"arg1": utils.Pointer("value1"),
				"arg2": utils.Pointer(""),
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
