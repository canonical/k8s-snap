package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
)

func TestClusterConfigEmpty(t *testing.T) {
	g := NewWithT(t)

	g.Expect(types.ClusterConfig{}.Empty()).To(BeTrue())
}
