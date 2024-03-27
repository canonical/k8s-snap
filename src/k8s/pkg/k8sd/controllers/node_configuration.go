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
	"github.com/mitchellh/mapstructure"
	v1 "k8s.io/api/core/v1"
)

type NodeConfigurationController struct {
	snap snap.Snap
}

func NewNodeConfigurationController(snap snap.Snap) *NodeConfigurationController {
	return &NodeConfigurationController{
		snap: snap,
	}
}

func (c *NodeConfigurationController) createK8sClient(ctx context.Context) *k8s.Client {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(3 * time.Second):
		default:
		}

		client, err := k8s.NewClient(c.snap.KubernetesNodeRESTClientGetter("kube-system"))
		if err != nil {
			// This fails when node is neither bootstrapped or joined
			continue
		}
		return client
	}
}

func (c *NodeConfigurationController) Run(ctx context.Context) {
	client := c.createK8sClient(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		default:
		}

		if err := client.WatchConfigMap(ctx, "kube-system", "k8sd-config", func(configMap *v1.ConfigMap) error { return c.reconcile(ctx, configMap) }); err != nil {
			// This also can fail during bootstrapping/start up when api-server is not ready
			// So the watch requests get connection refused replies
			log.Println(fmt.Errorf("error while watching configmap: %w", err))
		}
	}
}

func (c *NodeConfigurationController) reconcile(ctx context.Context, configMap *v1.ConfigMap) error {
	var nodeConfig types.NodeConfig
	if err := mapstructure.Decode(configMap.Data, &nodeConfig); err != nil {
		return fmt.Errorf("failed to decode node config: %w", err)
	}

	kubeletUpdateMap := make(map[string]string)
	var kubeletDeleteList []string

	if nodeConfig.ClusterDNS != "" {
		kubeletUpdateMap["--cluster-dns"] = nodeConfig.ClusterDNS
	} else {
		kubeletDeleteList = append(kubeletDeleteList, "--cluster-dns")
	}

	if nodeConfig.ClusterDomain != "" {
		kubeletUpdateMap["--cluster-domain"] = nodeConfig.ClusterDomain
	}

	mustRestartKubelet, err := snaputil.UpdateServiceArguments(c.snap, "kubelet", kubeletUpdateMap, kubeletDeleteList)
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
