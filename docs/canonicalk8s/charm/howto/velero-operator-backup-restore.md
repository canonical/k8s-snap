# Backup and restore using the Velero Operator

The [Velero Operator][] is a Juju Kubernetes charm that deploys and manages
[Velero][], an open-source tool for safely backing up and restoring Kubernetes
cluster resources. It provides disaster recovery, cluster migration, and backup
and restore for workloads across namespaces, including non-Juju-managed ones.

This guide shows how to setup Velero Operator with the [s3-integrator charm][]
as the storage provider and with the [infra-backup-operator] to backup the
configuration of **any** kind of Kubernetes distribution (Canonical Kubernetes,
MicroK8s, EKS, etc.).

## What the infra backup operator does
The Infra Backup Operator is a Juju charm designed to work seamlessly with the
Velero Operator. When related, it automatically applies the necessary
configuration to enable backups of Kubernetes resources that are
**not tied to workloads**, but to the cluster’s own configuration and
infrastructure. The backup is separated into two groups:

- Cluster Infra Backup -> All cluster-scoped resources.
- Namespaced infra backup -> All namespaced resources for Security and
Access Control (Role, RoleBinding, NetworkPolicy, etc.) and Configuration
and Environment (ConfigMap, Secret, etc.)

Note that because Kubernetes clusters might have different storage providers,
the infra-backup-operator does not create backup of PVs or PVCs.

## What you will need
- A kubernetes cluster
- A bootstrapped K8s controller. See the [Juju documentation]
- An S3 bucket or a S3 compatible bucket like [MinIO] or [microceph]

### Deploy
```bash
juju add-model velero

juju deploy velero-operator --trust
juju deploy infra-backup-operator
juju deploy s3-integrator
```

### Integrate
```bash
juju integrate velero-operator s3-integrator
juju integrate infra-backup-operator:cluster-infra-backup velero-operator
juju integrate infra-backup-operator:namespaced-infra-backup velero-operator
```

### Create the Backup
At any time users can run a juju action to create a backup
```bash
juju run velero-operator/0 create-backup target=infra-backup-operator:cluster-infra-backup
juju run velero-operator/0 create-backup target=infra-backup-operator:namespaced-infra-backup
```

### Restore
In case of disaster recovery, users can restore the cluster configuration in
the same cluster or in a different one using the Velero operator juju-action.
This will guarantee that the cluster configuration can be easily restored to
start receiving the workloads.

Before restore your cluster must have velero-operator deployed and integrated
with the same bucket of the backup.

```bash
# example output
juju run velero-operator/0 list-backups

backups:
  83503892-a24a-409b-b0df-553dcc2465ec:
    app: infra-backup-operator
    completion-timestamp: "2025-08-08T20:00:28Z"
    endpoint: cluster-infra-backup
    model: test-charm-9f0e8dda
    name: infra-backup-operator-cluster-infra-backup-pblz2
    phase: Completed
    start-timestamp: "2025-08-08T20:00:26Z"
  85662948-8e5e-4922-8e1c-c5568eafa6e7:
    app: infra-backup-operator
    completion-timestamp: "2025-08-07T18:42:13Z"
    endpoint: cluster-infra-backup
    model: test-charm-9f0e8dda
    name: infra-backup-operator-cluster-infra-backup-4bm7p
    phase: Completed
    start-timestamp: "2025-08-07T18:42:10Z"

# restore the backups
juju run velero-operator/0 restore backup-uid=85662948-8e5e-4922-8e1c-c5568eafa6e7
juju run velero-operator/0 restore backup-uid=83503892-a24a-409b-b0df-553dcc2465ec
```

## Workloads
Each Juju application should be responsible for setting the relation with Velero operator to
be able to backup the necessary k8s resources and in the right order.

<!-- Links -->

[Velero Operator]: https://charmhub.io/velero-operator
[Velero]: https://velero.io/
[s3-integrator charm]: https://charmhub.io/s3-integrator
[infra-backup-operator]: https://charmhub.io/infra-backup-operator/docs/tutorial
[Juju documentation]: https://documentation.ubuntu.com/juju/3.6/reference/juju-cli/list-of-juju-cli-commands/bootstrap/
[MinIO]: https://min.io/
[microceph]: https://canonical-microceph.readthedocs-hosted.com/stable/tutorial/get-started/
