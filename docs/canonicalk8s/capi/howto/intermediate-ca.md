# How to use intermediate CAs with Vault

By default, the ClusterAPI provider will generate self-signed CA certificates
for the workload clusters.

Follow this guide to prepare an intermediate Certificate Authority (CA) using
HashiCorp Vault and then configure ClusterAPI to use the generated certificates.

## Preparing Vault

For the purpose of this guide, we are going to install HashiCorp Vault using
snap and start a Vault server in development mode.

```bash
sudo snap install vault
vault server -dev &
```

Specify the Vault address through an environment variable:

```bash
export VAULT_ADDR=http://localhost:8200
```

Enable the PKI secrets engine and set the maximum lease time to 10 years
(87600 hours):

```bash
vault secrets enable pki
vault secrets tune -max-lease-ttl=87600h pki
```

## Generating the CA certificates

Generate the root CA certificate:

```bash
vault write pki/root/generate/internal \
    common_name=vault \
    ttl=87600h \
```

Generate the intermediate CA certificate. We need the resulting Certificate
Signing Request (CSR) and private key, so for convenience we'll use JSON
formatting and store the output in a file.

```bash
mkdir myca
vault write pki/intermediate/generate/exported common_name=kubernetes \
    -format=json > myca/intermediate.json
```

Extract the CSR and key to separate files:

```bash
cat myca/intermediate.json | jq -r '.data.csr' > myca/intermediate.csr
cat myca/intermediate.json | jq -r '.data.private_key' > myca/intermediate.key
```

Sign the intermediate CA using the root CA:

```bash
vault write -format=json pki/root/sign-intermediate \
    common_name=kubernetes \
    csr=@myca/intermediate.csr \
    ttl=87600h > myca/intermediate-signed.json
```

Extract the resulting intermediate CA certificate:

```bash
cat myca/intermediate-signed.json | jq -r '.data.ca_chain' \
    > myca/intermediate-chain.crt
cat myca/intermediate-signed.json | jq -r '.data.certificate' \
    > myca/intermediate.crt
```

## Passing intermediate CA certificates to CAPI

The Cluster API provider expects the CA certificates to be specified as
Kubernetes secrets named ``${cluster-name}-${purpose}``, where ``cluster-name``
is the name of the workload cluster and purpose is one of the following:


| Purpose suffix     | Description             |
|--------------------|-------------------------|
| ``ca``             | API server CA           |
| ``cca``            | client CA               |
| ``etcd``           | etcd CA (if using etcd) |
| ``proxy``          | Front Proxy CA          |

The secrets must have ``Opaque`` type, containing the ``tls.crt`` and
``tls.key`` fields.

Let's assume that we want to bootstrap a workload cluster named ``mycluster`` and
use the newly generated intermediate CA certificate. We'd first create the
following secret on the management cluster:

```bash
workloadClusterName="mycluster"
cat <<EOF > myca/ca-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: $workloadClusterName-ca
type: Opaque
stringData:
  tls.crt: |
$(cat myca/intermediate.crt | sed 's/^/    /g')
  tls.key: |
$(cat myca/intermediate.key | sed 's/^/    /g')
EOF

kubectl apply -f myca/ca-secret.yaml
```
