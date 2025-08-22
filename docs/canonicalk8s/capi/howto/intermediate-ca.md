# How to use intermediate CAs with Vault

By default, the ClusterAPI provider will generate self-signed CA certificates
for the workload clusters.

Follow this guide to prepare an intermediate Certificate Authority (CA) using
[HashiCorp Vault] and then configure ClusterAPI to use the generated
certificates.

```{include} /_parts/common_vault_intermediate_ca.md
```

## Pass intermediate CA certificates to CAPI

The Cluster API provider expects the CA certificates to be specified as
Kubernetes secrets named ``${cluster-name}-${purpose}``, where ``cluster-name``
is the name of the workload cluster and ``purpose`` is one of the following:

| Purpose suffix     | Description             |
|--------------------|-------------------------|
| ``ca``             | API server CA           |
| ``cca``            | client CA               |
| ``etcd``           | etcd CA (if using etcd) |
| ``proxy``          | Front Proxy CA          |

The secrets must have ``Opaque`` type, containing the ``tls.crt`` and
``tls.key`` fields.

Let's assume that we want to bootstrap a workload cluster named ``mycluster``
and use the newly generated intermediate CA certificate. We'd first create the
following secret on the management cluster:

```
cat <<EOF > myca/ca-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mycluster-ca
type: Opaque
stringData:
  tls.crt: |
$(cat myca/intermediate.crt | sed 's/^/    /g')
  tls.key: |
$(cat myca/intermediate.key | sed 's/^/    /g')
EOF
```

Now apply the CA certificate:

```
kubectl apply -f myca/ca-secret.yaml
```

The workload cluster will now retrieve the CA from the secret during
bootstrapping. Refer to the [provisioning guide] for further instructions on
completing the process.

## Further reading

See this [Vault article] for more details on how to integrate Vault as a
Kubernetes certificate manager.

<!--LINKS -->
[HashiCorp Vault]: https://developer.hashicorp.com/vault/docs
[provisioning guide]: ./provision.md
[Vault article]: https://support.hashicorp.com/hc/en-us/articles/21920341210899-Create-an-Intermediate-CA-in-Kubernetes-using-Vault-as-a-certificate-manager
