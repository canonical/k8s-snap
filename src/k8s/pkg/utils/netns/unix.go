package netnsutils

import (
	"context"

	mountutils "github.com/canonical/k8s/pkg/utils/mount"
)

type UnixNetworkNamespaceManager struct{}

func (h UnixNetworkNamespaceManager) ForEachNetworkNamespace(ctx context.Context, callback func(ctx context.Context, namespace string) error) error {
	return forEachNetworkNamespace(ctx, "/run/netns", callback)
}

func (h UnixNetworkNamespaceManager) DeleteNetworkNamespace(ctx context.Context, namespace string) error {
	mountHelper := mountutils.NewUnixMountHelper()
	return deleteNetworkNamespace(ctx, mountHelper, "/run/netns", namespace)
}

func NewUnixNetworkNSHelper() NetworkNamespaceManager {
	return &UnixNetworkNamespaceManager{}
}
