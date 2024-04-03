package controllers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/k8s"
	v1 "k8s.io/api/core/v1"
)

type NodeConfigurationController struct {
	snap            snap.Snap
	createK8sClient func(ctx context.Context) *k8s.Client
}

func NewNodeConfigurationController(snap snap.Snap, createK8sClient func(ctx context.Context) *k8s.Client) *NodeConfigurationController {
	return &NodeConfigurationController{
		snap:            snap,
		createK8sClient: createK8sClient,
	}
}

func (c *NodeConfigurationController) Run(ctx context.Context) {
	client := c.createK8sClient(ctx)
	for {
		if err := client.WatchConfigMap(ctx, "kube-system", "k8sd-config", func(configMap *v1.ConfigMap) error { return c.reconcile(ctx, configMap) }); err != nil {
			// This also can fail during bootstrapping/start up when api-server is not ready
			// So the watch requests get connection refused replies
			log.Println(fmt.Errorf("error while watching configmap: %w", err))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (c *NodeConfigurationController) reconcile(ctx context.Context, configMap *v1.ConfigMap) error {
	config, err := types.KubeletFromConfigMap(configMap.Data)
	if err != nil {
		return fmt.Errorf("failed to parse configmap data to kubelet config: %w", err)
	}

	updateArgs := make(map[string]string)
	var deleteArgs []string

	if config.ClusterDNS == nil {
		deleteArgs = append(deleteArgs, "--cluster-dns")
	} else {
		updateArgs["--cluster-dns"] = config.GetClusterDNS()
	}
	if config.ClusterDomain == nil {
		deleteArgs = append(deleteArgs, "--cluster-domain")
	} else {
		updateArgs["--cluster-domain"] = config.GetClusterDomain()
	}
	if config.CloudProvider == nil {
		deleteArgs = append(deleteArgs, "--cloud-provider")
	} else {
		updateArgs["--cloud-provider"] = config.GetCloudProvider()
	}

	mustRestartKubelet, err := snaputil.UpdateServiceArguments(c.snap, "kubelet", updateArgs, deleteArgs)
	if err != nil {
		return fmt.Errorf("failed to update kubelet arguments: %w", err)
	}

	if mustRestartKubelet {
		if err := c.snap.RestartService(ctx, "kubelet"); err != nil {
			return fmt.Errorf("failed to restart kubelet to apply node configuration: %w", err)
		}
	}

	return nil
}
