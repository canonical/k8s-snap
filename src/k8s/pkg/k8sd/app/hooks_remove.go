package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/log"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/microcluster/v3/cluster"
	"github.com/canonical/microcluster/v3/state"
)

// NOTE(ben): the pre-remove performs a series of cleanup steps on a best-effort basis.
// If any step fails, the error is logged, and the cleanup continues, skipping dependent tasks.
// All steps need to be blocking as the context is cancelled after the hook returned.
func (a *App) onPreRemove(ctx context.Context, s state.State, force bool) (rerr error) {
	snap := a.Snap()

	log := log.FromContext(ctx).WithValues("hook", "preremove", "node", s.Name())
	log.Info("Running preremove hook")

	log.Info("Waiting for node to finish microcluster join before removing")
	// NOTE (hue): in microcluster v2, PreRemove hook is also called if something goes wrong on
	// `bootstrap` and `join-cluster`. It is possible that we get stuck in this loop forever which causes
	// the `bootstrap` and `join-cluster` commands to hang and finally return an uninformative `context deadline exceeded` error
	// we optimistically stop trying after a fixed number of retries.
	maxRetries := 10
	var txnRetries int
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		var notPending bool
		if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			member, err := cluster.GetCoreClusterMember(ctx, tx, s.Name())
			if err != nil {
				log.Error(err, "Failed to get member")
				return nil
			}
			notPending = member.Role != cluster.Pending
			return nil
		}); err != nil {
			log.Error(err, "Failed database transaction to check cluster member role")
			txnRetries++
		}

		if txnRetries >= maxRetries {
			log.Info("Reached maximum number of retries for database transactions on pre-remove hook, continuing cleanup", "max_retries", maxRetries)
			return true, nil
		}

		return notPending, nil
	}); err != nil {
		log.Error(err, "Failed to wait for node to finish microcluster join before removing. Continuing with the cleanup...")
	}

	cfg, err := databaseutil.GetClusterConfig(ctx, s)
	if err == nil {
		if _, ok := cfg.Annotations.Get(apiv1_annotations.AnnotationSkipCleanupKubernetesNodeOnRemove); !ok {
			c, err := snap.KubernetesClient("")
			if err != nil {
				log.Error(err, "Failed to create Kubernetes client", err)
			}

			if c != nil {
				log.Info("Deleting node from Kubernetes cluster")
				if err := c.DeleteNode(ctx, s.Name()); err != nil {
					log.Error(err, "Failed to remove Kubernetes node")
				}
			}
		}

		switch cfg.Datastore.GetType() {
		case "k8s-dqlite":
			client, err := snap.K8sDqliteClient(ctx)
			if err == nil {
				log.Info("Removing node from k8s-dqlite cluster")
				nodeAddress := net.JoinHostPort(s.Address().Hostname(), fmt.Sprintf("%d", cfg.Datastore.GetK8sDqlitePort()))
				if err := client.RemoveNodeByAddress(ctx, nodeAddress); err != nil {
					// Removing the node might fail (e.g. if it is the only one in the cluster).
					// We still want to continue with the file cleanup, hence we only log the error.
					log.Error(err, "Failed to remove node from k8s-dqlite cluster")
				}
			} else {
				log.Error(err, "Failed to create k8s-dqlite client: %w")
			}

			log.Info("Cleaning up k8s-dqlite directory")
			if err := os.RemoveAll(snap.K8sDqliteStateDir()); err != nil {
				return fmt.Errorf("failed to cleanup k8s-dqlite state directory: %w", err)
			}
		case "external":
			log.Info("Cleaning up external datastore certificates")
			if _, err := setup.EnsureExtDatastorePKI(snap, &pki.ExternalDatastorePKI{}); err != nil {
				log.Error(err, "Failed to cleanup external datastore certificates")
			}
		default:
		}
	} else {
		log.Error(err, "Failed to retrieve cluster config")
	}

	for _, dir := range []string{snap.ServiceArgumentsDir()} {
		log.WithValues("directory", dir).Info("Cleaning up config files", dir)
		if err := os.RemoveAll(dir); err != nil {
			log.WithValues("dir", dir).Error(err, "failed to delete config files", err)
		}
	}

	// Perform all cleanup steps regardless of if this is a worker node or control plane.
	// Trying to detect the node type is not reliable as the node might have been marked as worker
	// or not, depending on which step it failed.
	log.Info("Cleaning up worker certificates")
	if _, err := setup.EnsureWorkerPKI(snap, &pki.WorkerNodePKI{}); err != nil {
		log.Error(err, "failed to cleanup worker certificates")
	}

	log.Info("Removing worker node mark")
	if err := snaputil.MarkAsWorkerNode(snap, false); err != nil {
		log.Error(err, "Failed to unmark node as worker")
	}

	log.Info("Cleaning up control plane certificates")
	if _, err := setup.EnsureControlPlanePKI(snap, &pki.ControlPlanePKI{}); err != nil {
		log.Error(err, "failed to cleanup control plane certificates")
	}

	if _, ok := cfg.Annotations.Get(apiv1.AnnotationSkipStopServicesOnRemove); !ok {
		log.Info("Stopping worker services")
		if err := snaputil.StopWorkerServices(ctx, snap); err != nil {
			log.Error(err, "Failed to stop worker services")
		}

		log.Info("Stopping control plane services")
		if err := snaputil.StopControlPlaneServices(ctx, snap); err != nil {
			log.Error(err, "Failed to stop control-plane services")
		}
	}

	return nil
}
