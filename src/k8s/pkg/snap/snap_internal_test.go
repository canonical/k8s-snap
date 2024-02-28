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
		{"With k8s. prefix", "k8s.test-service", "k8s.test-service"},
		{"Without prefix", "api", "k8s.api"},
		{"Just k8s", "k8s", "k8s"},
		{"Empty string", "", "k8s."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			got := serviceName(tc.input)
			g.Expect(got).To(Equal(tc.expected))
		})
	}
}
