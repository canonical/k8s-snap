# How to integrate {{product}} with ceph-csi

[Ceph] can be used to hold Kubernetes persistent volumes and is the recommended
storage solution for {{product}}.

The ``ceph-csi`` provisioner attaches the Ceph volumes to Kubernetes workloads.

## Prerequisites

This guide assumes an existing {{product}} cluster.
See the [charm installation] guide for more details.

In case of localhost/LXD Juju clouds, please make sure that the K8s units are
configured to use VM containers with Ubuntu 22.04 as the base and adding the
``virt-type=virtual-machine`` constraint. In order for K8s to function properly,
an adequate amount of resources must be allocated:

```{literalinclude} ../../_parts/install.md
:start-after: <!-- juju vm start -->
:end-before: <!-- juju vm end -->
```

## Deploy Ceph

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
```

If using LXD, configure the OSD unit to use VM containers by adding the
constraint: ``virt-type=virtual-machine``.

The `ceph-osd` and `ceph-mon` deployments should then be connected.

```
juju integrate ceph-osd:mon ceph-mon:osd
```

Once the units are ready, deploy ``ceph-csi``. By default, this enables
the ``ceph-xfs`` and ``ceph-ext4`` storage classes, which leverage
Ceph RBD.

```
CEPH_NS=ceph-ns  # kubernetes namespace for the ceph driver
juju deploy ceph-csi \
  --config provisioner-replicas=1 \
  --config namespace="${CEPH_NS}" \
  --config create-namespace=true
```

Integrate `ceph-csi` with our {{product}} cluster:

```
juju integrate ceph-csi k8s:ceph-k8s-info
juju integrate ceph-csi ceph-mon:client
```

`ceph-rbd` support is enabled by default but can be optionally disabled:

```
juju config ceph-csi ceph-rbd-enable=false
```

`ceph-fs` support can be optionally enabled (off by default):

```
juju deploy ceph-fs
juju integrate ceph-fs:ceph-mds ceph-mon:mds
juju config ceph-csi cephfs-enable=true
```

## Validate the CSI integration

Ensure that the storage classes are available and that the
CSI pods are running:

```
juju ssh k8s/leader -- sudo k8s kubectl get sc,po --namespace ${CEPH_NS}
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

So far, this guide demonstrates how to integrate with a single Ceph cluster.
However, {{product}} supports multiple Ceph clusters. The same `ceph-mon`,
`ceph-osd`, and `ceph-csi` charms can be deployed again as separate Juju
applications with different names.


```{note}
The alternate Ceph drivers will need a new namespace and resource names in the
deployment.
* Failure to configure a unique namespace will result in namespace collisions.
* Failure to configure each formatter could result in resource collisions.
```

Deploy an alternate Ceph cluster containing one monitor and one storage unit
(OSDs) -- again limiting the resources allocated.

In this example, we have provided the names `ceph-mon-alt` and `ceph-osd-alt`
for the Ceph cluster components to avoid collisions.

```
juju deploy -n 1 ceph-mon-alt ceph-mon \
    --constraints "cores=2 mem=4G root-disk=16G" \
    --config monitor-count=1 \
    --config expected-osd-count=1
juju deploy -n 1 ceph-osd-alt ceph-osd \
    --constraints "cores=2 mem=4G root-disk=16G" \
    --storage osd-devices=1G,1 --storage osd-journals=1G,1
```

Deploy `ceph-csi` again with a unique name, in this example `ceph-csi-alt`. We
have also provided a unique namespace for the additional cluster, `ceph-ns-alt`

Choose a name for the StorageClass Kubernetes Resources - `ext4`, `xfs` and
`cephfs`. They are managed by the `ceph-csi` charm and are cluster-wide
resources that have no namespace. For each of the supported StorageClass types,
there is an independent formatter:

* `ext4`, see [ceph-ext4-storage-class-name-formatter]
* `xfs`, see [ceph-xfs-storage-class-name-formatter]
* `cephfs`, see [cephfs-storage-class-name-formatter]

Each formatter has similar, but distinct formatting rules, so take care to plan
the storage-class names accordingly.

Finally, choose a name for the RBAC Kubernetes Resources. They are managed by
the `ceph-csi` charm and are cluster-wide resources that have no namespace.
For example `ClusterRole` and `ClusterRoleBinding`.

See [ceph-rbac-name-formatter] docs for details on choosing a name for RBAC
resources names.

```
CEPH_NS_ALT=ceph-ns-alt  # kubernetes namespace for the alternate ceph driver
juju deploy ceph-csi-alt ceph-csi \
    --config create-namespace=true \
    --config namespace=${CEPH_NS_ALT} \
    --config ceph-xfs-storage-class-name-formatter="ceph-xfs-{app}" \
    --config ceph-ext4-storage-class-name-formatter="ceph-ext4-{app}" \
    --config cephfs-storage-class-name-formatter="cephfs-{name}-{app}" \
    --config ceph-rbac-name-formatter="{name}-{app}"
    --config provisioner-replicas=1
```

Integrate all the new Ceph components with our {{product}} cluster:

```
juju integrate ceph-csi-alt k8s:ceph-k8s-info
juju integrate ceph-csi-alt ceph-mon-alt:client
juju integrate ceph-osd-alt:mon ceph-mon-alt:osd
```

## Resolve collisions

There are some Kubernetes Resources which can collide when deploying multiple
Ceph clusters in the same Kubernetes cluster if the names are not
specified correctly. If collisions occur, the new CephCSI application (`ceph-csi-alt`
in our example) enters a blocked state with status detailing the resource
conflicts it detects. For example:

```
10 Kubernetes resource collisions (action: list-resources)
```

List the collisions by running an action on the charm:

```
juju run ceph-csi-alt/leader list-resources
```

Resolve namespace collisions by ensuring the configuration for the two Ceph
drivers (`ceph-csi` and `ceph-csi-alt` in our example) are configured in
separate namespaces.

```
CEPH_NS_ALT=ceph-ns-alt  # kubernetes namespace for the alternate ceph driver
juju config ceph-csi-alt namespace=${CEPH_NS_ALT} create-namespace=true
```

Mitigate Storage Class collisions by ensuring these resources are named in
accordance with the upstream formatting rules linked above. For example:

```
juju config ceph-csi-alt cephfs-storage-class-name-formatter="cephfs-{name}-{app}"
```

Resolve RBAC collisions by ensuring these resources are named in accordance with
the Ceph RBAC formatting rules linked above. For example:

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
