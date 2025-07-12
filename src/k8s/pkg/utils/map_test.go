package utils_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestDeepCopyMap(t *testing.T) {
	g := NewWithT(t)

	original := map[string]string{
		"a": "1",
		"b": "2",
	}

	copied := utils.DeepCopyMap(original)
	g.Expect(copied).To(Equal(original))
	// Ensure original remains unchanged
	copied["a"] = "changed"
	g.Expect(original["a"]).To(Equal("1"))
	g.Expect(copied["a"]).To(Equal("changed"))
}
