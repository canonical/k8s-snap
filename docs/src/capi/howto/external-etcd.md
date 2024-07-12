# CAPI External Etcd

To replace the built-in [dqlite][dqlite] database with an external etcd to manage the kubernetes state in the CAPI workload cluster follow this `how-to guide`.

## Prerequisites

To follow this guide, you will need:

- A CAPI management cluster
- Initialize ClusterAPI
- Installed CRDs

<!-- Change this to getting started or whatever we come up with -->
Please refer to the [Development Guide](development.md) for instructions.

## Etcd Certificates

This example shows how to create the certificates for a 3-node etcd cluster
with self-signed certificates.

To create the etcd certificates, use the `cfssl` tool. To install [cfssl][cfssl]
on `linux-amd64` run the following commands:

```
mkdir ~/bin
curl -s -L -o ~/bin/cfssl https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
curl -s -L -o ~/bin/cfssljson https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
chmod +x ~/bin/{cfssl,cfssljson}
export PATH=$PATH:~/bin
```

You need the following certificates in one directory:

- root ca certificate
- root ca key
- gencert configuration
- etcd ca for each etcd node

Replace the certificates directory with the path to the directory where you have the certificates.

```
export CERTS_DIR="$PWD"
cfssl gencert --initca=true "$CERTS_DIR/etcd-root-ca-csr.json" | cfssljson --bare "$CERTS_DIR/etcd-root-ca"

```

cfssl gencert \
  --ca "$CERTS_DIR/etcd-root-ca.pem" \
  --ca-key "$CERTS_DIR/etcd-root-ca-key.pem" \
  --config "$CERTS_DIR/etcd-gencert.json" \
  "$CERTS_DIR/etcd-1-ca-csr.json" | cfssljson --bare "$CERTS_DIR/etcd-1"


## Run Etcd services via docker-compose

To run the etcd services via docker-compose, run:

```
export CERTS_DIR=/replace/me
docker-compose -f /replace/me/docker-compose.yaml up 
```


## Create Secrets

Create kubernetes secrets for the etcd certificates:

<!-- TODO: put these in a namespace? -->

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
  --from-file=ca.crt="$CERTS_DIR/etcd-root-ca.pem"
```

Create the secret for the etcd client:

```
kubectl create secret tls c1-apiserver-etcd-client \
  --cert=$CERTS_DIR/etcd-1.pem --key=$CERTS_DIR/etcd-1-key.pem 
```
<!-- Why etcd-1 only? -->

To confirm the secrets are created, run:

```
kubectl get secrets
```

## Update Cluster Template

The new control plane resource `CK8sControlPlane` needs to be configured to
store the kubernetes state. This is done by editing the cluster template and
adding this additional configuration:

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
clusterctl generate cluster c1 --from ./templates/docker/cluster-template.yaml --kubernetes-version v1.30.1 > c1.yaml
```

Create the cluster:

```
kubectl create -f c1.yaml
```

To check the status of the cluster, run:

```
kubectl get cluster,machine,ck8scontrolplane,secrets
```

<!-- LINKS -->

[cfssl]: https://github.com/cloudflare/cfssl
[dqlite]: https://dqlite.io/
