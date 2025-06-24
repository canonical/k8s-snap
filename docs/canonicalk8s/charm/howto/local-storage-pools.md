# How to deploy multiple local storage pools using the Rawfile LocalPV charm

This guide will walk you through deploying the [Rawfile LocalPV] charm in your
{{ product }} cluster to create and manage multiple tiered storage pools,
enabling flexible and optimized storage for your workloads. By leveraging node
selectors and different storage classes managed by the Rawfile LocalPV charm,
you can efficiently allocate storage based on data access patterns and
performance requirements.

## Prerequisites

This guide assumes an existing {{product}} cluster. See the
[charm installation] guide for setup details.


## Deploy multiple storage pools using `rawfile-localpv`

This example creates two `rawfile-localpv` applications with different
configurations for storage tiers.

The **fast tier** targets nodes labeled `storagePool=fast` with storage mounted
at `/mnt/fast` backed by high-performance disks.

The **cold tier** targets nodes labeled `storagePool=cold` with storage mounted
at `/mnt/cold` for infrequently accessed data.

This deployment uses [node selectors] to schedule the CSI workloads on nodes
with the required storage. You'll need to label your nodes appropriately to
ensure correct scheduling of workloads using persistent volumes. Check the
[k8s][k8s node labels] and [k8s-worker][k8s-worker node labels] charm
configuration options for node labeling instructions.

Fast storage pool (`fast-storage.yaml`):

```yaml
fast-localpv:
  namespace: fast
  storage-class-name: fast-sc
  node-selector: storagePool=fast
  node-storage-path: /mnt/fast
```

Cold storage pool (`cold-storage.yaml`):

```yaml
cold-localpv:
  namespace: cold
  storage-class-name: cold-sc
  node-selector: storagePool=cold
  node-storage-path: /mnt/cold
```

Deploy both applications and integrate with the `k8s` application:

```bash
juju deploy rawfile-localpv fast-localpv --config ./fast-storage.yaml
juju integrate k8s fast-localpv

juju deploy rawfile-localpv cold-localpv --config ./cold-storage.yaml
juju integrate k8s cold-localpv
```

Monitor the installation progress by running the following command:

```bash
juju status --watch 1s
```

## Validate the storage pools

To validate the storage pools, deploy a test pod that requests a persistent
volume from the **fast tier**. Since these volumes are node-bound, include the
same `nodeSelector` label to ensure the pod schedules on the correct node.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: fast-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: fast-sc
---
apiVersion: v1
kind: Pod
metadata:
  name: fast-pod
spec:
  nodeSelector:
    storagePool: fast
  containers:
    - name: writer
      image: busybox
      command: ["/bin/sh", "-c"]
      args:
        - echo "Hello!" > /data/fast.txt && sleep 3600
      volumeMounts:
        - name: data
          mountPath: /data
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: fast-pvc
```

To check if your PVC is bound, run:

```
kubectl get pvc fast-pvc -o jsonpath='{.status.phase}'
```

If everything is set up correctly, this command should return `Bound`, meaning
the PVC has successfully attached to the provisioner.

Next, confirm that the pod has attached the volume by running:

```
kubectl get pod fast-pod
```

If the pod is running as expected, you'll see output similar to:

```
NAME       READY   STATUS    RESTARTS   AGE
fast-pod   1/1     Running   0          2m47s
```

To test other storage pools, like the cold storage pool, just repeat these
steps using the appropriate node selectors and storage classes for each pool.

<!-- LINKS -->
[charm installation]: ./charm
[k8s]: https://charmhub.io/k8s
[k8s node labels]: https://charmhub.io/k8s/configurations#node-labels
[k8s-worker]: https://charmhub.io/k8s-worker
[k8s-worker node labels]: https://charmhub.io/k8s-worker/configurations#node-labels
[Rawfile LocalPV]: https://charmhub.io/rawfile-localpv
[node selectors]: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
