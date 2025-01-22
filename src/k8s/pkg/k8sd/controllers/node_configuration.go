package controllers

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/control"
	v1 "k8s.io/api/core/v1"
)

type NodeConfigurationController struct {
	snap      snap.Snap
	waitReady func()
}

func NewNodeConfigurationController(snap snap.Snap, waitReady func()) *NodeConfigurationController {
	return &NodeConfigurationController{
		snap:      snap,
		waitReady: waitReady,
	}
}

func (c *NodeConfigurationController) Run(ctx context.Context, getRSAKey func(context.Context) (*rsa.PublicKey, error)) {
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "node-configuration"))
	log := log.FromContext(ctx)

	log.Info("Waiting for node to be ready")
	// wait for microcluster node to be ready
	c.waitReady()

	log.Info("Starting node configuration controller")

	for {
		client, err := getNewK8sClientWithRetries(ctx, c.snap)
		if err != nil {
			log.Error(err, "Failed to create a Kubernetes client")
		}

		if err := client.WatchConfigMap(ctx, "kube-system", "k8sd-config", func(configMap *v1.ConfigMap) error { return c.reconcile(ctx, configMap, getRSAKey) }); err != nil {
			// This also can fail during bootstrapping/start up when api-server is not ready
			// So the watch requests get connection refused replies
			log.WithValues("name", "k8sd-config", "namespace", "kube-system").Error(err, "Failed to watch configmap")
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
		// This may fail if other controllers try to restart the services at the same time, hence the retry.
		if err := control.RetryFor(ctx, 5, 5*time.Second, func() error {
			if err := c.snap.RestartService(ctx, "kubelet"); err != nil {
				return fmt.Errorf("failed to restart kubelet to apply node configuration: %w", err)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed after retry: %w", err)
		}
	}

	return nil
}
