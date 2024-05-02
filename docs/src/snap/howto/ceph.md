# How to use ceph

Distributed, redundant storage is a must-have when you want to develop reliable applications. Ceph is a storage solution which provides exactly that, and is built with distributed clusters in mind. In this tutorial, we'll be looking at how to integrate Canonical Kubernetes with a Ceph cluster. Specifically, by the end of this tutorial you'll have a Kubernetes pod with a mounted RBD-backed volume. This how to is adapted from [block-devices-and-kubernetes].

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstrapped Canonical Kubernetes cluster (see the
  [getting-started-guide])
- You have a running Ceph cluster

## Create a Ceph storage pool

Create a storage pool named "kubernetes" in the Ceph cluster.

```
ceph osd pool create kubernetes
```

Initialize the pool as a Ceph block device (RBD) pool

```
rbd pool init kubernetes
```

## Configure ceph-csi

Ceph CSI is the CSI driver for Ceph. With Ceph CSI, Kubernetes will be able to accomplish tasks related to your Ceph cluster (like attaching volumes to workloads.)

Create a user for Kubernetes and ceph-csi.

```
ceph auth get-or-create client.kubernetes mon 'profile rbd' osd 'profile rbd pool=kubernetes' mgr 'profile rbd pool=kubernetes'
```

```
[client.kubernetes]
	key = AQBh1TNmFYERJhAAf5yqP4Wnrb/u4yNGsBKZHA==
```

Note the generated key.

## Generate ceph-csi-configmap.yaml

First, get the fsid and the monitor addresses of your cluster.

```
sudo ceph mon dump
```

```
epoch 2
fsid 6d5c12c9-6dfb-445a-940f-301aa7de0f29
last_changed 2024-05-02T14:01:37.668679-0400
created 2024-05-02T14:01:35.010723-0400
min_mon_release 18 (reef)
election_strategy: 1
0: [v2:10.0.0.136:3300/0,v1:10.0.0.136:6789/0] mon.dev
dumped monmap epoch 2
```

Note the v1 IP (10.0.0.136:6789) and the fsid (fsid 6d5c12c9-6dfb-445a-940f-301aa7de0f29).

```
cat <<EOF > csi-config-map.yaml
---
apiVersion: v1
kind: ConfigMap
data:
  config.json: |-
    [
      {
        "clusterID": "fsid 6d5c12c9-6dfb-445a-940f-301aa7de0f29",
        "monitors": [
          "10.0.0.136:6789",
        ]
      }
    ]
metadata:
  name: ceph-csi-config
EOF
```

Apply the ConfigMap object

```
kubectl apply -f csi-config-map.yaml
```

Next, here is a quote from the [block-devices-and-kubernetes] page.

> Recent versions of ceph-csi also require an additional ConfigMap object to define Key Management Service (KMS) provider details. If KMS isn’t set up, put an empty configuration in a csi-kms-config-map.yaml file or refer to examples at https://github.com/ceph/ceph-csi/tree/master/examples/kms:
```
cat <<EOF > csi-kms-config-map.yaml
---
apiVersion: v1
kind: ConfigMap
data:
  config.json: |-
    {}
metadata:
  name: ceph-csi-encryption-kms-config
EOF
```
```
kubectl apply -f csi-kms-config-map.yaml
```

Then, we need another ConfigMap which will be stored inside a ceph.conf file in the CSI containers.

```
cat <<EOF > ceph-config-map.yaml
---
apiVersion: v1
kind: ConfigMap
data:
  ceph.conf: |
    [global]
    auth_cluster_required = cephx
    auth_service_required = cephx
    auth_client_required = cephx
  # keyring is a required key and its value should be empty
  keyring: |
metadata:
  name: ceph-config
EOF
```
```
kubectl apply -f ceph-config-map.yaml
```

## Generate ceph-csi cephx secret
```
cat <<EOF > ceph-config-map.yaml
---
apiVersion: v1
kind: ConfigMap
data:
  ceph.conf: |
    [global]
    auth_cluster_required = cephx
    auth_service_required = cephx
    auth_client_required = cephx
  # keyring is a required key and its value should be empty
  keyring: |
metadata:
  name: ceph-config
EOF
```
```
kubectl apply -f ceph-config-map.yaml
```

## Create the ceph-csi cephx secret

```
cat <<EOF > csi-rbd-secret.yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: csi-rbd-secret
  namespace: default
stringData:
  userID: kubernetes
  userKey: AQBh1TNmFYERJhAAf5yqP4Wnrb/u4yNGsBKZHA==
EOF
```

```
kubectl apply -f csi-rbd-secret.yaml
```

<!-- LINKS -->
[getting-started-guide]: ../tutorial/getting-started.md
[block-devices-and-kubernetes]: https://docs.ceph.com/en/latest/rbd/rbd-kubernetes/