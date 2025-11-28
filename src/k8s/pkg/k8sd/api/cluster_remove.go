package api

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations"
	"github.com/canonical/k8s/pkg/client/kubernetes"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/k8s/pkg/utils/node"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v2/cluster"
	"github.com/canonical/microcluster/v2/state"
)

// postClusterRemove handles requests to remove a node from the cluster.
// It will remove the node from etcd/k8s-dqlite, microcluster and from Kubernetes.
// If force is true, the node is removed on a best-effort basis even if it is not reachable.
func (e *Endpoints) postClusterRemove(s state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	req := apiv1.RemoveNodeRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	if req.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	log := log.FromContext(ctx).WithValues("name", req.Name, "force", req.Force)

	client, err := snap.KubernetesClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to create k8s client: %w", err))
	}

	// Cordon the node to prevent new pods from being scheduled
	log.Info("Cordoning node")
	if err := client.CordonNode(ctx, req.Name); err != nil {
		if !req.Force {
			response.SmartError(fmt.Errorf("failed to cordon node %q: %w", req.Name, err))
		} else {
			log.Error(err, "Failed to cordon node, but continuing due to force flag")
		}
	}

	// Drain the node by evicting (or eventually deleting) all pods
	log.Info("Draining node")
	if err := client.DrainNode(ctx, req.Name, kubernetes.DrainOpts{
		DeleteEmptydirData: true,
		Force:              true,
		GracePeriodSeconds: 10,
		IgnoreDaemonsets:   true,
		AllowDeletion:      true,
	}); err != nil {
		if !req.Force {
			response.SmartError(fmt.Errorf("failed to drain node %q: %w", req.Name, err))
		} else {
			log.Error(err, "Failed to drain node, but continuing due to force flag")
		}
	}

	cfg, err := databaseutil.GetClusterConfig(ctx, s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	isControlPlane, err := node.IsControlPlaneNode(ctx, s, req.Name)
	if err != nil {
		if req.Force {
			log.Error(err, "Failed to determine if node is control-plane, but continuing due to force=true")
		} else {
			return response.InternalError(fmt.Errorf("failed to determine if node is control-plane: %w", err))
		}
	}

	if _, ok := cfg.Annotations[apiv1_annotations.AnnotationSkipCleanupKubernetesNodeOnRemove]; !ok {
		log.Info("Remove node from Kubernetes cluster")
		if err := removeNodeFromKubernetes(ctx, snap, req.Name); err != nil {
			if req.Force {
				// With force=true, we want to cleanup all out-of-sync mentions of this node.
				// It might be that the node is already gone from k8s, but not from microcluster.
				// So we log the error, but continue.
				log.Error(err, "Failed to remove node from Kubernetes, but continuing due to force=true")
			} else {
				return response.InternalError(fmt.Errorf("failed to remove node from Kubernetes: %w", err))
			}
		}
	} else {
		log.Info("Skipping Kubernetes node removal as per annotation")
	}

	// The control-plane check relies on the microcluster membership being correct. If the membership is out-of-sync, we might
	// mis-classify a control-plane node as a worker node. Hence we always proceed with the removal
	// if force=true, regardless of the role of the node.
	if isControlPlane || req.Force {
		log.Info("Remove node from datastore")
		if err := removeNodeFromDatastore(ctx, s, snap, req.Name, cfg); err != nil {
			if req.Force {
				// With force=true, we want to cleanup all out-of-sync mentions of this node.
				// So we log the error, but continue.
				log.Error(err, "Failed to remove node from datastore, but continuing due to force=true; ignore error for workers", "datastore", cfg.Datastore.GetType())
			} else {
				return response.InternalError(fmt.Errorf("failed to delete node from datastore: %w", err))
			}
		}

		log.Info("Remove node from microcluster")
		if err := removeNodeFromMicrocluster(ctx, s, req.Name, req.Force); err != nil {
			if req.Force {
				log.Error(err, "Failed to remove node from microcluster, but continuing due to force=true; ignore error for workers")
			} else {
				return response.InternalError(fmt.Errorf("failed to delete node from microcluster: %w", err))
			}
		}
	}

	return response.SyncResponse(true, &apiv1.RemoveNodeResponse{})
}

func removeNodeFromDatastore(ctx context.Context, s state.State, snap snap.Snap, nodeName string, clusterConfig types.ClusterConfig) error {
	switch clusterConfig.Datastore.GetType() {
	case "k8s-dqlite":
		if err := removeNodeFromK8sDqlite(ctx, s, snap, nodeName, clusterConfig); err != nil {
			return fmt.Errorf("failed to remove node from k8s-dqlite cluster: %w", err)
		}
	case "etcd":
		if err := removeNodeFromEtcd(ctx, snap, s, clusterConfig, nodeName); err != nil {
			return fmt.Errorf("failed to remove node from etcd cluster: %w", err)
		}
	case "external":
		// The admin is responsible for cleaning up the external datastore membership.
	default:
	}

	return nil
}

func removeNodeFromK8sDqlite(ctx context.Context, s state.State, snap snap.Snap, nodeName string, clusterConfig types.ClusterConfig) error {
	log := log.FromContext(ctx).WithValues("remove", "k8s-dqlite")

	client, err := snap.K8sDqliteClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create k8s-dqlite client: %w", err)
	}
	log.Info("Removing node from k8s-dqlite cluster")

	k8sClient, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	var nodeAddress string

	// k8s-dqlite does not have the concept of names, hence we need to check
	// in other places to resolve the nodeName to an address.
	// Since those could be diverged, we need to check multiple sources.
	// It first tries to get the node name from Kubernetes, then from microcluster database, and fails otherwise.
	node, err := k8sClient.GetNode(ctx, nodeName)
	if err != nil {
		log.Error(err, "Failed to get node from Kubernetes, falling back to microcluster database")
	} else {
		members, err := client.ListMembers(ctx)
		if err != nil {
			return fmt.Errorf("failed to list k8s-dqlite members: %w", err)
		}

		log.WithValues("kubernetes-members", node.Status.Addresses, "k8s-dqlite-members", members).Info("Matching Kubernetes node addresses with k8s-dqlite members")
		for _, addr := range node.Status.Addresses {
			for _, member := range members {
				host, _, err := net.SplitHostPort(member.Address)
				if err != nil {
					log.Error(err, "Failed to split host and port from k8s-dqlite member address", "address", member.Address)
					continue
				}
				if addr.Address == host {
					nodeAddress = addr.Address
					log.Info("Resolved node address from Kubernetes and matched with k8s-dqlite member", "address", nodeAddress)
					break
				}
			}
			if nodeAddress != "" {
				break
			}
		}

		if nodeAddress == "" {
			log.Error(fmt.Errorf("no match"), "Could not match Kubernetes node addresses with k8s-dqlite members")
		}
	}

	// Fall back to microcluster database if not found
	if nodeAddress == "" {
		err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			member, err := cluster.GetCoreClusterMember(ctx, tx, nodeName)
			if err != nil {
				return err
			}
			nodeAddress = member.Address
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to resolve node address from microcluster database: %w", err)
		}
		log.Info("Resolved node address from microcluster database", "address", nodeAddress)
	}

	if nodeAddress == "" {
		return fmt.Errorf("failed to resolve node address for node %q", nodeName)
	}

	// Remove port if present in the address
	if strings.Contains(nodeAddress, ":") {
		host, _, err := net.SplitHostPort(nodeAddress)
		if err != nil {
			return fmt.Errorf("failed to split host and port from node address: %w", err)
		}
		if host != "" {
			nodeAddress = host
		}
	}

	nodeAddress = net.JoinHostPort(nodeAddress, fmt.Sprintf("%d", clusterConfig.Datastore.GetK8sDqlitePort()))

	log.Info("Removing node from k8s-dqlite using address", "address", nodeAddress)
	if err := client.RemoveNodeByAddress(ctx, nodeAddress); err != nil {
		return fmt.Errorf("failed to remove node from k8s-dqlite cluster: %w", err)
	}
	return nil
}

func removeNodeFromEtcd(ctx context.Context, snap snap.Snap, s state.State, cfg types.ClusterConfig, nodeName string) error {
	leader, err := s.Leader()
	if err != nil {
		return fmt.Errorf("failed to get microcluster leader: %w", err)
	}
	members, err := leader.GetClusterMembers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get microcluster members: %w", err)
	}

	clientURLs := make([]string, 0, len(members)-1)
	for _, member := range members {
		if member.Name == nodeName {
			// skip the node we want to remove
			continue
		}
		clientURLs = append(clientURLs, fmt.Sprintf("https://%s", utils.JoinHostPort(member.Address.Addr().String(), cfg.Datastore.GetEtcdPort())))
	}

	client, err := snap.EtcdClient(clientURLs)
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %w", err)
	}
	defer client.Close()

	log := log.FromContext(ctx).WithValues("remove", "etcd", "name", nodeName, "clientURLs", clientURLs)
	log.Info("Deleting node from etcd cluster")
	if err := client.RemoveNodeByName(ctx, nodeName); err != nil {
		return fmt.Errorf("failed to remove node %s from etcd cluster: %w", nodeName, err)
	}

	return nil
}

func removeNodeFromMicrocluster(ctx context.Context, s state.State, nodeName string, force bool) error {
	log := log.FromContext(ctx).WithValues("name", nodeName)

	maxRetries := 10
	var txnRetries int
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		var notPending bool
		log.Info("Waiting for node to finish microcluster join before removing")
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
			log.Info("Reached maximum number of retries for database transactions", "max_retries", maxRetries)
			return true, nil
		}

		return notPending, nil
	}); err != nil {
		log.Error(err, "Failed to wait for node to finish microcluster join before removing. Continuing with the cleanup...")
	}

	// Remove control plane via microcluster API.
	c, err := s.Leader()
	if err != nil {
		return fmt.Errorf("failed to create client to cluster leader: %w", err)
	}

	// NOTE(hue): node removal process in CAPI might fail, we figured that the context passed to
	// `DeleteClusterMember` is somehow getting canceled but couldn't figure out why or by which component.
	// The cancellation happens after the `RunPreRemoveHook` call and before the `DeleteCoreClusterMember` call
	// in `clusterMemberDelete` endpoint of microcluster. This is a workaround to avoid the cancellation.
	// keep in mind that this failure is flaky and might not happen in every run.
	deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer deleteCancel()
	log.Info("Deleting node from Microcluster cluster, for real")
	if err := c.DeleteClusterMember(deleteCtx, nodeName, force); err != nil {
		return fmt.Errorf("failed to delete cluster member %s: %w", nodeName, err)
	}

	return nil
}

func removeNodeFromKubernetes(ctx context.Context, snap snap.Snap, nodeName string) error {
	log := log.FromContext(ctx)

	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}

	log.Info("Deleting node from Kubernetes cluster")
	if err := client.DeleteNode(ctx, nodeName); err != nil {
		return fmt.Errorf("failed to remove k8s node %q: %w", nodeName, err)
	}

	return nil
}
