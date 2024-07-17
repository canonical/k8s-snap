# CAPI External Etcd

To replace the built-in [dqlite][dqlite] database with an external etcd to
manage the kubernetes state in the CAPI workload cluster follow this
`how-to guide`.
This example shows how to create a 3-node workload cluster with an external
etcd. 

## Prerequisites

To follow this guide, you will need:

- Clusterctl
- A CAPI management cluster
- Initialize ClusterAPI
- Installed CRDs
- Secured 3-node etcd deployment

Please refer to the [getting-started][getting-started] for instructions.

## Create Kubernetes Secrets

Create three kubernetes secrets:

- {cluster-name}-etcd-servers
- {cluster-name}-etcd
- {cluster-name}-apiserver-etcd-client

```{note}
Replace {cluster-name} with the name of your cluster.
It is important to follow this naming convention since the providers will be looking for these names.
Please note that this example uses `c1` for the cluster name.
```

Create the secret for the etcd servers:

```
kubectl apply -f - <<EOF 
apiVersion: v1
kind: Secret
metadata:
  name: c1-etcd-servers
  namespace: default

data:
  servers: $(echo -n "https://etcd-1:2379,https://etcd-2:2379,https://etcd-3:2379" | base64 -w0)
EOF
```

To export the path to your etcd certs directory, use this command:

```
export CERTS_DIR=path/to/etcd-certs
```

Replace /path/to/etcd-certs with the actual path where you generated or stored
your etcd certificates.

Create the secret for the etcd root ca:

```
kubectl create secret generic c1-etcd \
  --from-file=tls.crt="$CERTS_DIR/etcd-root-ca.pem"
```

Create the `c1-apiserver-etcd-client` secret:

```
kubectl create secret tls c1-apiserver-etcd-client \
  --cert=$CERTS_DIR/etcd-1.pem --key=$CERTS_DIR/etcd-1-key.pem 
```

To confirm the secrets are created, run:

```
kubectl get secrets
```

## Etcd Cluster Template

The control plane resource `CK8sControlPlane` is configured to
store the kubernetes state in etcd. The cluster template `c1-external-etcd.yaml`
contains the following additional configuration:

```
controlPlane:
  datastoreType: external
  datastoreServersSecretRef:
    name: ${CLUSTER_NAME}-etcd-servers
    key: servers
```

## Deploy the Workload Cluster

To deploy the workload cluster, run:

```
export KIND_IMAGE=k8s-snap:dev
export CLUSTER_TEMPLATE_DIR=https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/assets
clusterctl generate cluster c1 --from $CLUSTER_TEMPLATE_DIR/c1-external-etcd.yaml --kubernetes-version v1.30.1 > c1.yaml
```

Create the cluster:

```
kubectl create -f c1.yaml
```

To check the status of the cluster, run:

```
clusterctl describe cluster c1 
```

## Optional: Delete workload cluster

To delete the workload cluster, run:

```bash
kubectl delete cluster c1
```

<!-- LINKS -->
[getting-started]: ../tutorial/getting-started.md
[cfssl]: https://github.com/cloudflare/cfssl
[dqlite]: https://dqlite.io/
[generate-etcd-certs]: https://raw.githubusercontent.com/canonical/k8s-snap/main/capi-ext-etcd/docs/src/assets/capi-etcd/generate-etcd-certs.sh
[capi-etcd]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/assets/capi-etcd/
