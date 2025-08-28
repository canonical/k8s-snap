package controllers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
)

// ControlPlaneConfigurationController watches for changes in the cluster configuration
// and applies them on the control plane services.
type ControlPlaneConfigurationController struct {
	snap      snap.Snap
	waitReady func()
	triggerCh <-chan time.Time
	// reconciledCh is used to notify that the controller has finished its reconciliation loop.
	reconciledCh chan struct{}
}

// NewControlPlaneConfigurationController creates a new controller.
// triggerCh is typically a `time.NewTicker(<duration>).C`.
func NewControlPlaneConfigurationController(snap snap.Snap, waitReady func(), triggerCh <-chan time.Time) *ControlPlaneConfigurationController {
	return &ControlPlaneConfigurationController{
		snap:         snap,
		waitReady:    waitReady,
		triggerCh:    triggerCh,
		reconciledCh: make(chan struct{}, 1),
	}
}

// Run starts the controller.
// Run accepts a context to manage the lifecycle of the controller.
// Run accepts a function that retrieves the current cluster configuration.
// Run will loop every time the trigger channel is.
func (c *ControlPlaneConfigurationController) Run(ctx context.Context, getClusterConfig func(context.Context) (types.ClusterConfig, error)) {
	c.waitReady()

	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "control-plane-configuration"))
	log := log.FromContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.triggerCh:
		}

		if isWorker, err := snaputil.IsWorker(c.snap); err != nil {
			log.Error(err, "Failed to check if running on a worker node")
			continue
		} else if isWorker {
			log.Info("Stopping on worker node")
			return
		}

		config, err := getClusterConfig(ctx)
		if err != nil {
			log.Error(err, "Failed to retrieve cluster configuration")
			continue
		}

		if err := c.reconcile(ctx, config); err != nil {
			log.Error(err, "Failed to reconcile control plane configuration")
		}

		select {
		case c.reconciledCh <- struct{}{}:
		default:
		}
	}
}

func (c *ControlPlaneConfigurationController) reconcile(ctx context.Context, config types.ClusterConfig) error {
	// kube-apiserver: external datastore
	if config.Datastore.GetType() == "external" {
		// certificates
		certificatesChanged, err := setup.EnsureExtDatastorePKI(c.snap, &pki.ExternalDatastorePKI{
			DatastoreCACert:     config.Datastore.GetExternalCACert(),
			DatastoreClientCert: config.Datastore.GetExternalClientCert(),
			DatastoreClientKey:  config.Datastore.GetExternalClientKey(),
		})
		if err != nil {
			return fmt.Errorf("failed to reconcile external datastore certificates: %w", err)
		}

		// kube-apiserver arguments
		updateArgs, deleteArgs, err := config.Datastore.ToKubeAPIServerArguments(c.snap)
		if err != nil {
			return fmt.Errorf("failed to get datastore arguments for kube-apiserver: %w", err)
		}

		argsChanged, err := snaputil.UpdateServiceArguments(c.snap, "kube-apiserver", updateArgs, deleteArgs)
		if err != nil {
			return fmt.Errorf("failed to update kube-apiserver datastore arguments: %w", err)
		}

		if certificatesChanged || argsChanged {
			if err := c.snap.RestartServices(ctx, []string{"kube-apiserver"}); err != nil {
				return fmt.Errorf("failed to restart kube-apiserver to apply configuration: %w", err)
			}
		}
	} else if config.Datastore.GetType() == "k8s-dqlite" {
		// Get existing feature gates to merge our changes
		featureGates, err := snaputil.GetServiceArgument(c.snap, "kube-apiserver", "--feature-gates")
		if err != nil {
			return fmt.Errorf("failed to get kube-apiserver feature gates: %w", err)
		}

		// (KU-4140): Look into improving UpdateServiceArguments to handle more complex values
		// instead of treating everything as a simple string
		featureGatesMap := make(map[string]bool)
		if featureGates != "" {
			pairs := strings.Split(featureGates, ",")
			for _, pair := range pairs {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					gate := strings.TrimSpace(kv[0])
					stringVal := strings.TrimSpace(kv[1])
					boolVal, err := strconv.ParseBool(stringVal)
					if err != nil {
						return fmt.Errorf("failed to parse feature gate value %q for key %q: %w", stringVal, kv[0], err)
					}
					featureGatesMap[gate] = boolVal
				}
			}
		}

		// (KU-4139): Disable feature gates incompatible with k8s-dqlite introduced in kubernetes 1.34
		featureGatesMap["ListFromCacheSnapshot"] = false
		featureGatesMap["SizeBasedListCostEstimate"] = false
		featureGatesMap["DetectCacheInconsistency"] = false

		var featureGatesList []string
		for k, v := range featureGatesMap {
			featureGatesList = append(featureGatesList, fmt.Sprintf("%s=%t", k, v))
		}
		args := map[string]string{"--feature-gates": strings.Join(featureGatesList, ",")}
		mustRestart, err := snaputil.UpdateServiceArguments(c.snap, "kube-apiserver", args, nil)
		if err != nil {
			return fmt.Errorf("failed to render arguments file: %w", err)
		}

		if mustRestart {
			// This may fail if other controllers try to restart the services at the same time, hence the retry.
			if err := control.RetryFor(ctx, 5, 5*time.Second, func() error {
				if err := c.snap.RestartServices(ctx, []string{"kube-apiserver"}); err != nil {
					return fmt.Errorf("failed to restart kube-apiserver to apply feature gates: %w", err)
				}
				return nil
			}); err != nil {
				return fmt.Errorf("failed after retry: %w", err)
			}
		}
	}

	// kube-controller-manager: cloud-provider
	if v := config.Kubelet.CloudProvider; v != nil {
		mustRestart, err := snaputil.UpdateServiceArguments(c.snap, "kube-controller-manager", map[string]string{"--cloud-provider": *v}, nil)
		if err != nil {
			return fmt.Errorf("failed to update kube-controller-manager arguments: %w", err)
		}

		if mustRestart {
			if err := c.snap.RestartServices(ctx, []string{"kube-controller-manager"}); err != nil {
				return fmt.Errorf("failed to restart kube-controller-manager to apply configuration: %w", err)
			}
		}
	}

	// snapd
	if meta, _, err := snapdconfig.ParseMeta(ctx, c.snap); err == nil && meta.Orb != "none" {
		if err := snapdconfig.SetSnapdFromK8sd(ctx, config.ToUserFacing(), c.snap); err != nil {
			log.FromContext(ctx).Error(err, "Failed to update snapd configuration")
		}
	}

	return nil
}

// ReconciledCh returns the channel where the controller pushes when a reconciliation loop is finished.
func (c *ControlPlaneConfigurationController) ReconciledCh() <-chan struct{} {
	return c.reconciledCh
}
