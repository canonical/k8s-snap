package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NodeConfigurationReconciler struct {
	client.Client
	scheme *runtime.Scheme
	snap   snap.Snap

	waitReady func()

	getClusterConfig func(ctx context.Context) (types.ClusterConfig, error)

	reconciledCh chan struct{}
}

func NewNodeConfigurationReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	snap snap.Snap,
	waitReady func(),
) *NodeConfigurationReconciler {
	return &NodeConfigurationReconciler{
		Client:       client,
		scheme:       scheme,
		snap:         snap,
		waitReady:    waitReady,
		reconciledCh: make(chan struct{}, 1),
	}
}

func (c *NodeConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).WithEventFilter(predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isTargetConfigMap(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isTargetConfigMap(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isTargetConfigMap(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isTargetConfigMap(e.Object)
		},
	}).Complete(c)
}

func isTargetConfigMap(obj client.Object) bool {
	return obj.GetName() == "k8sd-config" && obj.GetNamespace() == "kube-system"
}

func (c *NodeConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues(
		"controller", "update-node-configuration",
		"configmap", req.NamespacedName,
	)

	// Check if we're running on a worker node
	if isWorker, err := snaputil.IsWorker(c.snap); err != nil {
		logger.Error(err, "Failed to check if running on a worker node")
		return reconcile.Result{RequeueAfter: time.Second * 30}, err
	} else if isWorker {
		logger.Info("Running on worker node, skipping reconciliation")
		return reconcile.Result{}, nil
	}

	// Get cluster configuration
	config, err := c.getClusterConfig(ctx)
	if err != nil {
		logger.Error(err, "Failed to retrieve cluster configuration")
		return reconcile.Result{RequeueAfter: time.Second * 30}, err
	}

	// Load and process certificates
	keyPEM := config.Certificates.GetK8sdPrivateKey()
	key, err := pkiutil.LoadRSAPrivateKey(keyPEM)
	if err != nil && keyPEM != "" {
		return reconcile.Result{}, fmt.Errorf("failed to load cluster RSA key: %w", err)
	}

	// Generate ConfigMap data
	cmData, err := config.Kubelet.ToConfigMap(key)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to format kubelet configmap data: %w", err)
	}

	// Get existing ConfigMap
	cm := &corev1.ConfigMap{}
	if err := c.Get(ctx, req.NamespacedName, cm); err != nil {
		logger.Error(err, "Failed to get ConfigMap")
		return reconcile.Result{}, err
	}

	// Update ConfigMap
	cm.Data = cmData
	if err := c.Update(ctx, cm); err != nil {
		logger.Error(err, "Failed to update ConfigMap")
		return reconcile.Result{}, err
	}

	// Notify that reconciliation is complete
	select {
	case c.reconciledCh <- struct{}{}:
	default:
	}

	return reconcile.Result{}, nil
}

// ReconciledCh returns the channel that receives notifications when reconciliation completes
func (c *NodeConfigurationReconciler) ReconciledCh() <-chan struct{} {
	return c.reconciledCh
}

func (r *NodeConfigurationReconciler) SetConfigGetter(getter func(context.Context) (types.ClusterConfig, error)) {
	r.getClusterConfig = getter
}
