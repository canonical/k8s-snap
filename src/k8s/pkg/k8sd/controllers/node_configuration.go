package controllers

import (
	"context"
	"crypto/rsa"
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
	snap         snap.Snap
	waitReady    func()
	newK8sClient func() (*k8s.Client, error)
}

func NewNodeConfigurationController(snap snap.Snap, waitReady func(), newK8sClient func() (*k8s.Client, error)) *NodeConfigurationController {
	return &NodeConfigurationController{
		snap:         snap,
		waitReady:    waitReady,
		newK8sClient: newK8sClient,
	}
}

func (c *NodeConfigurationController) retryNewK8sClient(ctx context.Context) (*k8s.Client, error) {
	for {
		client, err := c.newK8sClient()
		if err == nil {
			return client, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

func (c *NodeConfigurationController) Run(ctx context.Context, getRSAKey func(context.Context) (*rsa.PublicKey, error)) {
	// wait for microcluster node to be ready
	c.waitReady()

	for {
		client, err := c.retryNewK8sClient(ctx)
		if err != nil {
			log.Println(fmt.Errorf("failed to create a Kubernetes client: %w", err))
		}

		if err := client.WatchConfigMap(ctx, "kube-system", "k8sd-config", func(configMap *v1.ConfigMap) error { return c.reconcile(ctx, configMap, getRSAKey) }); err != nil {
			// This also can fail during bootstrapping/start up when api-server is not ready
			// So the watch requests get connection refused replies
			log.Println(fmt.Errorf("failed to watch configmap: %w", err))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (c *NodeConfigurationController) reconcile(ctx context.Context, configMap *v1.ConfigMap, getRSAKey func(context.Context) (*rsa.PublicKey, error)) error {
	key, err := getRSAKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to load the RSA public key: %w", err)
	}
	config, err := types.KubeletFromConfigMap(configMap.Data, key)
	if err != nil {
		return fmt.Errorf("failed to parse configmap data to kubelet config: %w", err)
	}

	updateArgs := make(map[string]string)
	var deleteArgs []string

	for _, loop := range []struct {
		val *string
		arg string
	}{
		{arg: "--cloud-provider", val: config.CloudProvider},
		{arg: "--cluster-dns", val: config.ClusterDNS},
		{arg: "--cluster-domain", val: config.ClusterDomain},
	} {
		switch {
		case loop.val == nil:
			// value is not set in the configmap, no-op
		case *loop.val == "":
			// value is set in the configmap to the empty string, delete argument
			deleteArgs = append(deleteArgs, loop.arg)
		case *loop.val != "":
			// value is set in the configmap, update argument
			updateArgs[loop.arg] = *loop.val
		}
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
