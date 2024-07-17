# Backup and restore

[Velero][] is a popular open source backup solution for Kubernetes. Its core
implementation is a controller running in the cluster that oversees the backup
and restore operations. The administrator is given a CLI tool to schedule
operations and/or perform on-demand backup and restores. This CLI tool creates
Kubernetes resources that the in-cluster Velero controller acts upon. During
installation the controller needs to be [configured with a repository (called a
'provider')][providers], where the backup files are stored.

This document describes how to setup Velero with the MinIO provider acting as
an S3 compatible object store.


## What you will need

- A running Canonical Kubernetes with DNS enabled
- MinIO (install described below) or other S3 bucket
- Velero (install described below)
- An example workload

### Enabling required components

DNS is needed for this setup:

```bash
sudo k8s enable dns
```

### Install MinIO

[MinIO][] provides an S3 compatible interface over storage provisioned by
Kubernetes. For the purposes of this guide, the `local storage` component is
used to satisfy the persistent volume claims:

```bash
sudo k8s enable local-storage
```

Helm is used to setup [MinIO][MinIO Charts] under the `velero` namespace:

```bash
sudo k8s helm repo add minio https://charts.min.io
sudo k8s kubectl create namespace velero

# PRODUCTION MODE
sudo k8s helm install -n velero \
--set buckets[0].name=velero,buckets[0].policy=none,buckets[0].purge=false \
--generate-name minio/minio
```

### Create a demo workload

The workload we will demonstrate the backup with is an NGINX deployment and a
corresponding service under the `workloads` namespace. Create this setup with:

```bash
sudo k8s kubectl create namespace workloads
sudo k8s kubectl create deployment nginx -n workloads --image nginx
sudo k8s kubectl expose deployment nginx -n workloads --port 80
```


## Install Velero

Download the Velero binary from the 
[releases page on github][releases] and place it in our `PATH`. In this case we
install the v1.14.1 Linux binary for AMD64 under `/usr/local/bin`:

```bash
wget https://github.com/vmware-tanzu/velero/releases/download/v1.14.0/velero-v1.14.0-linux-amd64.tar.gz 
tar -xzf velero-v1.14.0-linux-amd64.tar.gz
chmod +x velero-v1.14.0-linux-amd64/velero
sudo chown root:root velero-v1.14.0-linux-amd64/velero
sudo mv velero-v1.14.0-linux-amd64/velero /usr/local/bin/velero
```

Before installing Velero into the cluster, we export the kubeconfig file using
the `config` command.

```bash
mkdir -p $HOME/.kube
sudo k8s kubectl config view --raw > $HOME/.kube/config
```

We also export the MinIO credentials so we can provide them to Velero. Be aware
that MinIO is used as an S3 bucket replacement for AWS S3, hence the
credentials look like they require aws values. This is merely the nomenclature
for accessing the S3 bucket in MinIO.

```bash
ACCESS_KEY=$(sudo k8s kubectl -n velero get secret -l app=minio -o jsonpath="{.items[0].data.rootUser}" | base64 --decode)
SECRET_KEY=$(sudo k8s kubectl -n velero get secret -l app=minio -o jsonpath="{.items[0].data.rootPassword}" | base64 --decode)
SERVICE=$(sudo k8s kubectl -n velero get svc -l app=minio -o jsonpath="{.items[0].metadata.name}")
cat <<EOF > credentials-velero
[default]
    aws_access_key_id=${ACCESS_KEY}
    aws_secret_access_key=${SECRET_KEY}
EOF
```

We are now ready to install Velero into the cluster, with an aws plugin that
[matches][aws-plugin-matching] the velero release:

```bash
SERVICE_URL="http://${SERVICE}.velero.svc:9000"
BUCKET=velero
REGION=minio
velero install \
--provider aws \
--plugins velero/velero-plugin-for-aws:v1.10.0 \
--bucket $BUCKET \
--backup-location-config region=$REGION,s3ForcePathStyle="true",s3Url=$SERVICE_URL \
--snapshot-location-config region=$REGION \
--secret-file ./credentials-velero
```


## Backup workloads

To backup the `workloads` namespace we use the `--include-namespaces` argument:
 
```bash
velero backup create workloads-backup --include-namespaces=workloads
```

```{note} Please see the 
[official Velero documentation](https://velero.io/docs/v1.14/file-system-backup/#to-back-up)
for details on how to backup persistent volumes and the supported volume types.
```

To check the progress of a backup operation we use `describe`, providing the
backup name:

```bash
velero backup describe workloads-backup 
```

In the output you should see this operation completed: 

```bash
Name:         workloads-backup
Namespace:    velero
Labels:       velero.io/storage-location=default
Annotations:  velero.io/resource-timeout=10m0s
              velero.io/source-cluster-k8s-gitversion=v1.30.2
              velero.io/source-cluster-k8s-major-version=1
              velero.io/source-cluster-k8s-minor-version=30

Phase:  Completed


Namespaces:
  Included:  workloads
  Excluded:  <none>

Resources:
  Included:        *
  Excluded:        <none>
  Cluster-scoped:  auto

Label selector:  <none>

Or label selector:  <none>

Storage Location:  default

Velero-Native Snapshot PVs:  auto
Snapshot Move Data:          false
Data Mover:                  velero

TTL:  720h0m0s

CSISnapshotTimeout:    10m0s
ItemOperationTimeout:  4h0m0s

Hooks:  <none>

Backup Format Version:  1.1.0

Started:    2024-07-15 19:35:18 +0000 UTC
Completed:  2024-07-15 19:35:18 +0000 UTC

Expiration:  2024-08-14 19:35:18 +0000 UTC

Total items to be backed up:  18
Items backed up:              18

Backup Volumes:
  Velero-Native Snapshots: <none included>

  CSI Snapshots: <none included>

  Pod Volume Backups: <none included>

HooksAttempted:  0
HooksFailed:     0
```


## Restore workloads

Before restoring the workloads namespace, let's delete it first:

```bash
sudo k8s kubectl delete namespace workloads
```

We can now create a restore operation specifying the backup we want to use:

```bash
velero restore create --from-backup workloads-backup
```

A restore operation which we can monitor using the describe command is then
created:

```bash
velero restore describe workloads-backup-20240715193553
```

The `describe` output should eventually report a “Completed” phase:

```bash
Name:         workloads-backup-20240715193553
Namespace:    velero
Labels:       <none>
Annotations:  <none>

Phase:                       Completed
Total items to be restored:  11
Items restored:              11

Started:    2024-07-15 19:35:54 +0000 UTC
Completed:  2024-07-15 19:35:54 +0000 UTC

Warnings:
  Velero:     <none>
  Cluster:  could not restore, CustomResourceDefinition "ciliumendpoints.cilium.io" already exists. Warning: the in-cluster version is different than the backed-up version
  Namespaces: <none>

Backup:  workloads-backup

Namespaces:
  Included:  all namespaces found in the backup
  Excluded:  <none>

Resources:
  Included:        *
  Excluded:        nodes, events, events.events.k8s.io, backups.velero.io, restores.velero.io, resticrepositories.velero.io, csinodes.storage.k8s.io, volumeattachments.storage.k8s.io, backuprepositories.velero.io
  Cluster-scoped:  auto

Namespace mappings:  <none>

Label selector:  <none>

Or label selector:  <none>

Restore PVs:  auto

CSI Snapshot Restores: <none included>

Existing Resource Policy:   <none>
ItemOperationTimeout:       4h0m0s

Preserve Service NodePorts:  auto

Uploader config:


HooksAttempted:   0
HooksFailed:      0
```

Listing the resources of the `workloads` namespaces confirms that the
restoration process was successful:

```bash
sudo k8s kubectl get all -n workloads
```

## Summing up

Although Velero is a really powerful tool with a large set of configuration
options it is also very easy to use. You are required to set up a backup
strategy based on the backend that will hold the backups and the scheduling of
the backups. The rest is taken care of by the tool itself.

    
<!-- Links -->

[Velero]: https://velero.io/
[providers]: https://velero.io/docs/v1.14/supported-providers/
[MinIO]: https://min.io/
[MinIO Charts]:  https://charts.min.io/
[releases]: https://github.com/vmware-tanzu/velero/releases
[aws-plugin-matching]: https://github.com/vmware-tanzu/velero-plugin-for-aws?tab=readme-ov-file#compatibility