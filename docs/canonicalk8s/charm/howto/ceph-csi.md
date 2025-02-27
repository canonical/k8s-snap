# How to integrate {{product}} with ceph-csi

[Ceph] can be used to hold Kubernetes persistent volumes and is the recommended
storage solution for {{product}}.

The ``ceph-csi`` provisioner attaches the Ceph volumes to Kubernetes workloads.

## Prerequisites

This guide assumes that you have an existing {{product}} cluster.
See the [charm installation] guide for more details.

In case of localhost/LXD Juju clouds, please make sure that the K8s units are
configured to use VM containers with Ubuntu 22.04 as the base and adding the
``virt-type=virtual-machine`` constraint. In order for K8s to function properly,
an adequate amount of resources must be allocated:

```{literalinclude} ../../_parts/install.md
:start-after: <!-- juju vm start -->
:end-before: <!-- juju vm end -->
```

## Deploying Ceph

Deploy a Ceph cluster containing one monitor and one storage unit
(OSDs). In this example, a limited amount of resources is being allocated.

```
juju deploy -n 1 ceph-mon \
    --constraints "cores=2 mem=4G root-disk=16G" \
    --config monitor-count=1 \
    --config expected-osd-count=1
juju deploy -n 1 ceph-osd \
    --constraints "cores=2 mem=4G root-disk=16G" \
    --storage osd-devices=1G,1 --storage osd-journals=1G,1
juju integrate ceph-osd:mon ceph-mon:osd
```

If using LXD, configure the OSD unit to use VM containers by adding the
constraint: ``virt-type=virtual-machine``.

Once the units are ready, deploy ``ceph-csi``. By default, this enables
the ``ceph-xfs`` and ``ceph-ext4`` storage classes, which leverage
Ceph RBD.

```
juju deploy ceph-csi --config provisioner-replicas=1
juju integrate ceph-csi k8s:ceph-k8s-info
juju integrate ceph-csi ceph-mon:client
```

CephFS support can optionally be enabled:

```
juju deploy ceph-fs
juju integrate ceph-fs:ceph-mds ceph-mon:mds
juju config ceph-csi cephfs-enable=True
```

## Validating the CSI integration

Ensure that the storage classes are available and that the
CSI pods are running:

```
juju ssh k8s/leader -- sudo k8s kubectl get sc,po --namespace default
```

The list should include the ``ceph-xfs`` and ``ceph-ext4`` storage classes as
well as ``cephfs``, if it was enabled.

Verify that Ceph PVCs work as expected. Connect to the k8s leader unit
and define a PVC like so:

```
juju ssh k8s/leader

cat <<EOF > /tmp/pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: raw-block-pvc
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 64Mi
  storageClassName: ceph-xfs
EOF

sudo k8s kubectl apply -f /tmp/pvc.yaml
```

Next, create a pod that writes to a Ceph volume:

```
cat <<EOF > /tmp/writer.yaml
apiVersion: v1
kind: Pod
metadata:
  name: pv-writer-test
  namespace: default
spec:
  restartPolicy: Never
  volumes:
  - name: pvc-test
    persistentVolumeClaim:
      claimName: raw-block-pvc
  containers:
  - name: pv-writer
    image: busybox
    command: ["/bin/sh", "-c", "echo 'PVC test data.' > /pvc/test_file"]
    volumeMounts:
    - name: pvc-test
      mountPath: /pvc
EOF

sudo k8s kubectl apply -f /tmp/writer.yaml
```

If the pod completes successfully, our Ceph CSI integration is functional.

```
sudo k8s kubectl wait pod/pv-writer-test \
    --for=jsonpath='{.status.phase}'="Succeeded" \
    --timeout 2m
```

## Relate to multiple Ceph clusters

So far, this guide demonstrates to how to integrate with a single ceph cluster
represented by the single `ceph-mon` application. But {{product}} supports
multiple ceph clusters via the same set of charms deployed again as different
applications.

Deploy an alternate Ceph cluster containing one monitor and one storage unit
(OSDs). Again, lets limit the resources allocated.

```
juju deploy -n 1 ceph-mon-alt ceph-mon \
    --constraints "cores=2 mem=4G root-disk=16G" \
    --config monitor-count=1 \
    --config expected-osd-count=1
juju deploy -n 1 ceph-osd-alt ceph-osd \
    --constraints "cores=2 mem=4G root-disk=16G" \
    --storage osd-devices=1G,1 --storage osd-journals=1G,1
juju deploy ceph-csi-alt ceph-csi \
    --config provisioner-replicas=1
juju integrate ceph-csi-alt k8s:ceph-k8s-info
juju integrate ceph-csi-alt ceph-mon-alt:client
juju integrate ceph-osd-alt:mon ceph-mon-alt:osd
```

These applications still uses the same charms, but represent new application
instances.  A new ceph-cluster via `ceph-mon-alt` and `ceph-osd-alt` and a new
integration with kubernetes by `ceph-csi-alt`.

There are some Kubernetes Resources which collide in this deployment style
and the admin will notice the `ceph-csi-alt` charm blocks because those conflict
are detected.

The `ceph-csi-alt` unit may end up in the `blocked` state with a status:

example)
`10 Kubernetes resource collisions (action: list-resources)`

By running an action on the charm, we can list those collisions:

```
juju run ceph-csi-alt/leader list-resources
```

### Resolving collisions

#### Namespace collisions

Many of the Kubernetes Resources managed by the `ceph-csi` charm have an 
associated namespace. Let's make sure we adjust the config for the
`ceph-csi-alt` application so that it doesn't collide with `ceph-csi`.

```
juju exec k8s/leader -- k8s kubectl create namespace ceph-csi-alt
juju config ceph-csi-alt namespace=ceph-csi-alt
```

After this we should see the number of collisions between the two applications
drop off, but there could still be collisions to investigate.

#### Storage Class collisions

StorageClass Kubernetes Resources managed by the `ceph-csi` charm are
cluster-wide resources and have no namespace.

For each of the supported storage class types, there is an independent formatter
for that class type.

For `ext4`, see [ceph-ext4-storage-class-name-formatter]
For `xfs`, see [ceph-xfs-storage-class-name-formatter]
For `cephfs`, see [cephfs-storage-class-name-formatter]

Each formatter has similar but potentially distinct formatting rules, so
take time to plan your storage-class names accordingly.

example)

```
juju config ceph-csi-alt cephfs-storage-class-name-formatter="cephfs-{name}-{app}"
```

#### RBAC collisions

RBAC Kubernetes Resources managed by the `ceph-csi` charm are cluster-wide
resources and have no namespace. Two such resources are `ClusterRole` and
`ClusterRoleBinding`.

The charm can be configured to craft separate names for these resources.  The
juju admin can format the names of these objects using a custom formatter.

See [ceph-rbac-name-formatter] docs for more details.

```
juju config ceph-csi-alt ceph-rbac-name-formatter="{name}-{app}"
```

<!-- LINKS -->

[charm installation]: ./charm
[Ceph]: https://docs.ceph.com/
[ceph-rbac-name-formatter]: https://charmhub.io/ceph-csi/configurations?channel=latest/edge#ceph-rbac-name-formatter
[ceph-ext4-storage-class-name-formatter]: https://charmhub.io/ceph-csi/configurations?channel=latest/edge#ceph-ext4-storage-class-name-formatter
[ceph-xfs-storage-class-name-formatter]: https://charmhub.io/ceph-csi/configurations?channel=latest/edge#ceph-xfs-storage-class-name-formatter
[cephfs-storage-class-name-formatter]: https://charmhub.io/ceph-csi/configurations?channel=latest/edge#cephfs-storage-class-name-formatter
