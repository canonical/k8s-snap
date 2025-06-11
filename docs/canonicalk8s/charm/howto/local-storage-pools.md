# How to deploy multiple local storage pools with Rawfile LocalPV

This guide will walk you through deploying [Rawfile LocalPV] in your
{{ product }} cluster to create multiple storage pools across your
nodes.

## Prerequisites

This guide assumes an existing {{product}} cluster. See the
[charm installation] guide for setup details.

```{important}
This deployment uses [node selectors] to schedule the CSI workloads on nodes
with the required storage. You'll need to label your nodes appropriately to
ensure correct scheduling of workloads using persistent volumes. Check the
[k8s] and [k8s-worker] charm configuration options for node labeling
instructions.
```

The examples below demonstrate the deployment pattern. Adjust node
configurations, labels and parameters based on your infrastructure needs.
Review the [k8s], [k8s-worker], and [rawfile-localpv][Rawfile LocalPV] pages
for all available customization options.

## Deploy multiple storage pools using `rawfile-localpv`

This example creates two `rawfile-localpv` applications with different
configurations for storage tiers.

The **fast tier** targets nodes labeled `storagePool=fast` with storage mounted
at `/mnt/fast` backed by high-performance disks.

The **cold tier** targets nodes labeled `storagePool=cold` with storage mounted
at `/mnt/cold` for infrequently accessed data.

1. Fast storage pool (`fast-storage.yaml`):

```yaml
fast-localpv:
  namespace: fast
  storage-class-name: fast-sc
  node-selector: storagePool=fast
  node-storage-path: /mnt/fast
```

2. Storage pool for cold storage (`cold-storage.yaml`):

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

<!-- LINKS -->
[charm installation]: ./charm
[k8s]: https://charmhub.io/k8s
[k8s-worker]: https://charmhub.io/k8s-worker
[Rawfile LocalPV]: https://charmhub.io/rawfile-localpv
[node selectors]: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector
