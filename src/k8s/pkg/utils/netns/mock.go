package netnsutils

import (
	"context"

	mountutils "github.com/canonical/k8s/pkg/utils/mount"
)

type MockNetworkNamespaceManager struct {
	netnsDir string
}

func (h MockNetworkNamespaceManager) ForEachNetworkNamespace(ctx context.Context, callback func(ctx context.Context, namespace string) error) error {
	return forEachNetworkNamespace(ctx, h.netnsDir, callback)
}

func (h MockNetworkNamespaceManager) DeleteNetworkNamespace(ctx context.Context, namespace string) error {
	mountHelper := mountutils.MockMountManager{}
	return deleteNetworkNamespace(ctx, mountHelper, h.netnsDir, namespace)
}

func NewMockNetworkNSHelper(netnsDir string) NetworkNamespaceManager {
	return &MockNetworkNamespaceManager{
		netnsDir: netnsDir,
	}
}
