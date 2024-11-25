# How to integrate {{product}} with ceph-csi

[Ceph] can be used to hold Kubernetes persistent volumes and is the recommended
storage solution for {{product}}.

The ``ceph-csi`` plugin automatically provisions and attaches the Ceph volumes
to Kubernetes workloads.

## What you'll need

This guide assumes that you have an existing {{product}} cluster.
See the [charm installation] guide for more details.

In case of localhost/LXD Juju clouds, please make sure that the K8s units are
configured to use VM containers with Ubuntu 22.04 as the base and adding the 
``virt-type=virtual-machine`` constraint. In order for K8s to function properly,
an adequate amount of resources must be allocated:

```
juju deploy k8s --channel=latest/edge \
    --base "ubuntu@22.04" \
    --constraints "cores=2 mem=8G root-disk=16G virt-type=virtual-machine"
```

## Deploying Ceph

Deploy a Ceph cluster containing one monitor and three storage units
(OSDs). In this example, a limited amount of reources is being allocated.

```
juju deploy -n 1 ceph-mon \
    --constraints "cores=2 mem=4G root-disk=16G" \
    --config monitor-count=1
juju deploy -n 3 ceph-osd \
    --constraints "cores=2 mem=4G root-disk=16G" \
    --storage osd-devices=1G,1 --storage osd-journals=1G,1
juju integrate ceph-osd:mon ceph-mon:osd
```

If using LXD, configure the OSD units to use VM containers by adding the 
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

<!-- LINKS -->

[charm installation]: ./charm
[Ceph]: https://docs.ceph.com/

