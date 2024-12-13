package app

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/controllers"
	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	"github.com/canonical/microcluster/v2/state"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func (a *App) onStart(ctx context.Context, s state.State) error {
	// start a goroutine to mark the node as running
	go func() {
		if err := a.markNodeReady(ctx, s); err != nil {
			log.FromContext(ctx).Error(err, "Failed to mark node as ready")
		}
	}()

	// start node config controller
	if a.nodeConfigController != nil {
		go a.nodeConfigController.Run(ctx, func(ctx context.Context) (*rsa.PublicKey, error) {
			cfg, err := databaseutil.GetClusterConfig(ctx, s)
			if err != nil {
				return nil, fmt.Errorf("failed to load RSA key from configuration: %w", err)
			}
			keyPEM := cfg.Certificates.GetK8sdPublicKey()
			key, err := pkiutil.LoadRSAPublicKey(cfg.Certificates.GetK8sdPublicKey())
			if err != nil && keyPEM != "" {
				return nil, fmt.Errorf("failed to load RSA key: %w", err)
			}
			return key, nil
		})
	}

	// start control plane config controller
	if a.controlPlaneConfigController != nil {
		go a.controlPlaneConfigController.Run(ctx, func(ctx context.Context) (types.ClusterConfig, error) {
			return databaseutil.GetClusterConfig(ctx, s)
		})
	}

	// start update node config controller
	if a.nodeConfigReconciler != nil {
		a.nodeConfigReconciler.SetConfigGetter(func(ctx context.Context) (types.ClusterConfig, error) {
			return databaseutil.GetClusterConfig(ctx, s)
		})
	}
	// start feature controller
	if a.featureController != nil {
		go a.featureController.Run(
			ctx,
			func(ctx context.Context) (types.ClusterConfig, error) {
				return databaseutil.GetClusterConfig(ctx, s)
			},
			func() (string, error) {
				c, err := s.Leader()
				if err != nil {
					return "", fmt.Errorf("failed to get leader client: %w", err)
				}

				clusterMembers, err := c.GetClusterMembers(ctx)
				if err != nil {
					return "", fmt.Errorf("failed to get cluster members: %w", err)
				}

				localhostAddress, err := DetermineLocalhostAddress(clusterMembers)
				if err != nil {
					return "", fmt.Errorf("failed to determine localhost address: %w", err)
				}

				return localhostAddress, nil
			},
			func(ctx context.Context, dnsIP string) error {
				if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					if _, err := database.SetClusterConfig(ctx, tx, types.ClusterConfig{
						Kubelet: types.Kubelet{ClusterDNS: utils.Pointer(dnsIP)},
					}); err != nil {
						return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
					}
					return nil
				}); err != nil {
					return fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
				}

				return nil
			},
			func(ctx context.Context, name types.FeatureName, featureStatus types.FeatureStatus) error {
				if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					// we set timestamp here in order to reduce the clutter. otherwise we will need to
					// set .UpdatedAt field in a lot of places for every event/error.
					// this is not 100% accurate but should be good enough
					featureStatus.UpdatedAt = time.Now()
					if err := database.SetFeatureStatus(ctx, tx, name, featureStatus); err != nil {
						return fmt.Errorf("failed to set feature status in db for %q: %w", name, err)
					}
					return nil
				}); err != nil {
					return fmt.Errorf("database transaction to set feature status failed: %w", err)
				}
				return nil
			},
		)
	}

	// start csrsigning controller
	if a.csrsigningController != nil {
		go a.csrsigningController.Run(
			ctx,
			func(ctx context.Context) (types.ClusterConfig, error) {
				return databaseutil.GetClusterConfig(ctx, s)
			},
		)
	}

	go func() {
		mgr, err := setupManager(ctx, a)
		if err != nil {
			log.FromContext(ctx).Error(err, "Failed to setup manager")
			return
		}

		a.manager = mgr
		log.FromContext(ctx).Info("Starting controller manager")
		if err := mgr.Start(ctx); err != nil {
			log.FromContext(ctx).Error(err, "Manager failed to start")
		}
	}()

	return nil
}

func setupManager(ctx context.Context, app *App) (manager.Manager, error) {
	log.FromContext(ctx).Info("Setting up controller manager, waiting for ready signal")
	app.readyWg.Wait()
	log.FromContext(ctx).Info("Received ready signal, setting up controller manager")

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	k8sClient, err := app.config.Snap.KubernetesClient("kube-system")
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	log.FromContext(ctx).Info("Created kubernetes client")

	options := ctrl.Options{
		Scheme: scheme,
	}

	mgr, err := ctrl.NewManager(k8sClient.RESTConfig(), options)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	log.FromContext(ctx).Info("Created controller manager")

	if _, err := setupControllers(ctx, app, mgr); err != nil {
		return nil, fmt.Errorf("failed to setup controllers: %w", err)
	}

	return mgr, nil
}

func setupControllers(ctx context.Context, app *App, mgr manager.Manager) (*controllers.NodeConfigurationReconciler, error) {
	log.FromContext(ctx).Info("Setting up controllers")
	if app.nodeConfigReconciler != nil {
		log.FromContext(ctx).Info("Setting up node configuration reconciler")

		app.nodeConfigReconciler.SetClient(mgr.GetClient())
		app.nodeConfigReconciler.SetScheme(mgr.GetScheme())

		if err := app.nodeConfigReconciler.SetupWithManager(mgr); err != nil {
			return nil, fmt.Errorf("failed to setup node configuration reconciler: %w", err)
		}
		return app.nodeConfigReconciler, nil
	}

	return nil, nil
}
