# Use external etcd with Cluster API

To replace the built-in datastore with an external etcd to
manage the Kubernetes state in the Cluster API (CAPI) workload cluster follow
this `how-to guide`. This example shows how to create a 3-node workload cluster
with an external etcd.

## Prerequisites

To follow this guide, you will need:

- [clusterctl][clusterctl] installed
- A CAPI management cluster initialised with the infrastructure, bootstrap and
  control plane providers of your choice. Please refer to the
  [getting-started guide][getting-started] for instructions.
- Secured 3-node etcd deployment

## Create Kubernetes secrets

Create three Kubernetes secrets:

- peaches-etcd-servers
- peaches-etcd
- peaches-apiserver-etcd-client

```{note}
Replace `peaches` with the name of your cluster.
It is important to follow this naming convention for the secrets since the providers will be looking for these names.
```

Create the secret for the etcd servers:

```
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: peaches-etcd-servers
  namespace: default

stringData:
  servers: https://etcd-1:2379,https://etcd-2:2379,https://etcd-3:2379
EOF
```

```{note}
Replace `https://etcd-1:2379,https://etcd-2:2379,https://etcd-3:2379` with the actual etcd server addresses.
```

To export the path to your etcd certs directory, use this command:

```
export CERTS_DIR=path/to/etcd-certs
```

Replace `/path/to/etcd-certs` with the actual path where you generated or stored
your etcd certificates.

Create the secret for the etcd root ca:

```
kubectl create secret generic peaches-etcd \
  --from-file=tls.crt="$CERTS_DIR/etcd-root-ca.pem"
```

Create the `peaches-apiserver-etcd-client` secret:

```
kubectl create secret tls peaches-apiserver-etcd-client \
  --cert=$CERTS_DIR/etcd-1.pem --key=$CERTS_DIR/etcd-1-key.pem
```

To confirm the secrets are created, run:

```
kubectl get secrets
```

## Update etcd cluster template

Please refer to [CAPI-templates][CAPI-templates] for the latest templates.
Update the control plane resource `CK8sControlPlane` so that it is configured to
store the Kubernetes state in etcd. Add the following additional configuration
to the cluster template `cluster-template.yaml`:

```
apiVersion: controlplane.cluster.x-k8s.io/v1beta2
kind: CK8sControlPlane
metadata:
  name: ${CLUSTER_NAME}-control-plane
spec:
  # ...
  spec:
    # ...
    controlPlane:
      datastoreType: external
      datastoreServersSecretRef:
        name: ${CLUSTER_NAME}-etcd-servers
        key: servers
```

## Deploy the workload cluster

To deploy the workload cluster, run:

```
clusterctl generate cluster peaches --from ./cluster-template.yaml --kubernetes-version v1.30.1 > peaches.yaml
```

Create the cluster:

```
kubectl create -f peaches.yaml
```

To check the status of the cluster, run:

```
clusterctl describe cluster peaches
```
<!-- LINKS -->
[getting-started]: ../tutorial/getting-started.md
[CAPI-templates]: https://github.com/canonical/cluster-api-k8s/tree/main/templates
[clusterctl]: https://cluster-api.sigs.k8s.io/clusterctl/overview
