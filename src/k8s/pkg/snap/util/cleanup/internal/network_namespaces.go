package internal

import (
	"context"
	"strings"

	"github.com/canonical/k8s/pkg/log"
	netnsutils "github.com/canonical/k8s/pkg/utils/netns"
)

func RemoveNetworkNamespaces(ctx context.Context, netnsHelper netnsutils.NetworkNamespaceManager) {
	log := log.FromContext(ctx)

	err := netnsHelper.ForEachNetworkNamespace(ctx, func(ctx context.Context, namespace string) error {
		if strings.HasPrefix(namespace, "cni-") {
			return netnsHelper.DeleteNetworkNamespace(ctx, namespace)
		}
		return nil
	})
	if err != nil {
		log.Error(err, "failed to iterate and delete network namespaces")
	}
}
