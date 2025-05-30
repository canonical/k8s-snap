package internal_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap/util/cleanup/internal"
	netnsutils "github.com/canonical/k8s/pkg/utils/netns"
	. "github.com/onsi/gomega"
)

func TestRemoveNetworkNamespaces(t *testing.T) {
	ctx := context.Background()
	netnsDir := t.TempDir()
	g := NewWithT(t)

	helper := netnsutils.NewMockNetworkNSHelper(netnsDir)

	// Create namespaces: one cni- and one non-cni-
	cniNS := "cni-test-ns"
	otherNS := "other-ns"
	cniPath := filepath.Join(netnsDir, cniNS)
	otherPath := filepath.Join(netnsDir, otherNS)
	g.Expect(os.Mkdir(cniPath, 0o755)).To(Succeed())
	g.Expect(os.Mkdir(otherPath, 0o755)).To(Succeed())
	g.Expect(cniPath).To(BeAnExistingFile())
	g.Expect(otherPath).To(BeAnExistingFile())

	internal.RemoveNetworkNamespaces(ctx, helper)

	// cni- namespace should be deleted, other should remain
	g.Expect(cniPath).ToNot(BeAnExistingFile())
	g.Expect(otherPath).To(BeAnExistingFile())
}
