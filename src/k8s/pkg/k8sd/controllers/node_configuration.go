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
}

func NewNodeConfigurationController() *NodeConfigurationController {
	return &NodeConfigurationController{}
}

func (c *NodeConfigurationController) Run(ctx context.Context, snap snap.Snap) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(3 * time.Second):
		default:
		}

		client, err := k8s.NewClient(snap.KubernetesNodeRESTClientGetter("kube-system"))
		if err != nil {
			// This fails when the node is not bootstrapped or joined
			log.Println(fmt.Errorf("failed to create kubernetes node client: %w", err))
			continue
		}

		if err := client.WatchConfigMap(ctx, "kube-system", "k8sd-config", func(configMap *v1.ConfigMap) error { return c.reconcile(ctx, snap, configMap) }); err != nil {
			// This also can fail during bootstrapping/start up when api-server is not ready
			// So the watch requests get connection refused replies
			log.Println(fmt.Errorf("error while watching configmap: %w", err))
		}
	}
}

func (c *NodeConfigurationController) reconcile(ctx context.Context, snap snap.Snap, configMap *v1.ConfigMap) error {
	var nodeConfig types.NodeConfig
	if err := mapstructure.Decode(configMap.Data, &nodeConfig); err != nil {
		return fmt.Errorf("failed to decode node config: %w", err)
	}

	var kubeletUpdateMap map[string]string = make(map[string]string)
	var kubeletDeleteList []string

	if nodeConfig.ClusterDNS != "" {
		kubeletUpdateMap["--cluster-dns"] = nodeConfig.ClusterDNS
	} else {
		kubeletDeleteList = append(kubeletDeleteList, "--cluster-dns")
	}

	if nodeConfig.ClusterDomain != "" {
		kubeletUpdateMap["--cluster-domain"] = nodeConfig.ClusterDomain
	}

	mustRestartKubelet, err := snaputil.UpdateServiceArguments(snap, "kubelet", kubeletUpdateMap, kubeletDeleteList)
	if err != nil {
		return fmt.Errorf("failed to update kubelet arguments: %w", err)
	}

	if mustRestartKubelet {
		if err := snap.RestartService(ctx, "kubelet"); err != nil {
			return fmt.Errorf("failed to restart kubelet to apply node configuration: %w", err)
		}
	}

	return nil
}
