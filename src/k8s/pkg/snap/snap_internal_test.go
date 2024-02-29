package snap

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestServiceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"WithPrefix", "k8s.test-service", "k8s.test-service"},
		{"NoPrefix", "api", "k8s.api"},
		{"K8s", "k8s", "k8s"},
		{"EmptyString", "", "k8s."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			got := serviceName(tc.input)
			g.Expect(got).To(Equal(tc.expected))
		})
	}
}
