package newtypes_test

import (
	"testing"

	newtypes "github.com/canonical/k8s/pkg/k8sd/new"
	. "github.com/onsi/gomega"
)

func TestClusterConfigEmpty(t *testing.T) {
	g := NewWithT(t)

	g.Expect(newtypes.ClusterConfig{}.Empty()).To(BeTrue())
}
