# Use external etcd with Cluster API

To replace the built-in datastore with an external etcd to
manage the Kubernetes state in the Cluster API (CAPI) workload cluster follow
this `how-to guide`. This example shows how to create a 3-node workload cluster
with an external etcd. 

## Prerequisites

To follow this guide, you will need:

- [Clusterctl][clusterctl] installed
- A CAPI management cluster initialised with the infrastructure, bootstrap and
  control plane providers of your choice
- Secured 3-node etcd deployment

Please refer to the [getting-started][getting-started] for instructions.

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

Replace /path/to/etcd-certs with the actual path where you generated or stored
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

Update the control plane resource `CK8sControlPlane` so that it is configured to
store the Kubernetes state in etcd. The cluster template `peaches.yaml` contains
the following additional configuration:

```
controlPlane:
  datastoreType: external
  datastoreServersSecretRef:
    name: ${CLUSTER_NAME}-etcd-servers
    key: servers
```

## Deploy the workload cluster

To deploy the workload cluster, run:

```
clusterctl generate cluster peaches --from peaches.yaml --kubernetes-version v1.30.1 > peaches.yaml
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
[capi-etcd]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/assets/capi-etcd/
[cfssl]: https://github.com/cloudflare/cfssl
[clusterctl]: https://cluster-api.sigs.k8s.io/clusterctl/overview
[dqlite]: https://dqlite.io/
[generate-etcd-certs]: https://raw.githubusercontent.com/canonical/k8s-snap/main/capi-ext-etcd/docs/src/assets/capi-etcd/generate-etcd-certs.sh
