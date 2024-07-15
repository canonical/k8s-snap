# CAPI External Etcd

To replace the built-in [dqlite][dqlite] database with an external etcd to
manage the kubernetes state in the CAPI workload cluster follow this
`how-to guide`.

## Prerequisites

To follow this guide, you will need:

- Clusterctl
- A CAPI management cluster
- Initialize ClusterAPI
- Installed CRDs

Please refer to the [getting-started][getting-started] for instructions.

## Copy capi-etcd directory

This example shows how to create a 3-node workload cluster with an external
etcd. To follow along with the example copy the [capi-etcd][capi-etcd]
directory from github. 

### Install cfssl

To install cfssl, follow the [upstream instructions][cfssl]. Typically, this
involves fetching the executable that matches your hardware architecture and
placing it in your PATH. For example, at the time this guide was written,
for `linux-amd64` you would run:

```
mkdir ~/bin
curl -s -L -o ~/bin/cfssl https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
curl -s -L -o ~/bin/cfssljson https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
chmod +x ~/bin/{cfssl,cfssljson}
export PATH=$PATH:~/bin
```

### Create Certificates

In the copied [capi-etcd][capi-etcd]
directory you will find the following certificate files:

- root ca certificate `etcd-root-ca-csr.json`
- root ca key `etcd-root-ca-csr.pem`
- gencert configuration `etcd-gencert.json`
- etcd ca for each etcd node `etcd-1-ca-csr.json`, `etcd-2-ca-csr.json`,
  `etcd-3-ca-csr.json`

Optionally, you can edit the certificates to match your requirements. Export
the directory where the certificates are stored as `EXT_ETCD_DIR` and run the
[generate-etc-certs][generate-etcd-certs] script:

```
export EXT_ETCD_DIR=/path/to/capi-etcd
$EXT_ETCD_DIR/generate-etcd-certs.sh
```

## Run Etcd services via docker-compose

To run the etcd services via docker-compose, run:

```
docker-compose -f $EXT_ETCD_DIR/docker-compose.yaml up 
```

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

Create the secret for the etcd root ca:

```
kubectl create secret generic c1-etcd \
  --from-file=ca.crt="$EXT_ETCD_DIR/etcd-root-ca.pem"
```

Create the `c1-apiserver-etcd-client` secret:

```
kubectl create secret tls c1-apiserver-etcd-client \
  --cert=$EXT_ETCD_DIR/etcd-1.pem --key=$EXT_ETCD_DIR/etcd-1-key.pem 
```
<!-- Why etcd-1 only? -->

To confirm the secrets are created, run:

```
kubectl get secrets
```

## Etcd Cluster Template

The new control plane resource `CK8sControlPlane` is configured to
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
clusterctl generate cluster c1 --from $EXT_ETCD_DIR/c1-external-etcd.yaml --kubernetes-version v1.30.1 > c1.yaml
```

Create the cluster:

```
kubectl create -f c1.yaml
```

To check the status of the cluster, run:

```
clusterctl describe cluster c1 
```

<!-- LINKS -->
[getting-started]: ../tutorial/getting-started.md
[cfssl]: https://github.com/cloudflare/cfssl
[dqlite]: https://dqlite.io/
[generate-etcd-certs]: https://raw.githubusercontent.com/canonical/k8s-snap/main/capi-ext-etcd/docs/src/assets/capi-etcd/generate-etcd-certs.sh
[capi-etcd]: https://raw.githubusercontent.com/canonical/k8s-snap/main/docs/src/assets/capi-etcd/
