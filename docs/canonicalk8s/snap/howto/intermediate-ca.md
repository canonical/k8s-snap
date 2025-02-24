# How to use intermediate CAs with Vault

By default, {{product}} will generate self-signed CA certificates for the
Kubernetes services.

Follow this guide to prepare an intermediate Certificate Authority (CA) using
[HashiCorp Vault] and then configure {{product}} to use the generated
certificates.

```{include} /_parts/common_vault_intermediate_ca.md
```

## Pass intermediate CA certificates to {{product}}

The CA certificates can be specified through the [bootstrap configuration file]
using the following fields:

| Field                  | Description                |
|------------------------|----------------------------|
| ``ca-crt``             | API server CA certificate  |
| ``ca-key``             | API server CA key          |
| ``client-ca-crt``      | client CA certificate      |
| ``client-ca-key``      | client CA key              |
| ``front-proxy-ca-crt`` | Front Proxy CA certificate |
| ``front-proxy-ca-key`` | Front Proxy CA key         |

Prepare a bootstrap configuration using our newly generated intermediate CA
certificate.

```
cat <<EOF > myca/bootstrap.yaml
ca-crt: |
$(cat myca/intermediate.crt | sed 's/^/    /g')
ca-key: |
$(cat myca/intermediate.key | sed 's/^/    /g')
cluster-config:
  network:
    enabled: true
  dns:
    enabled: true
  local-storage:
    enabled: true
EOF
```

Now bootstrap the cluster:

```
sudo k8s bootstrap --file myca/bootstrap.yaml
```

Use this command to wait for the cluster to become ready:

```
sudo k8s status --wait-ready
```

Check the following files to ensure that the expected CA certificates were
applied:

* ``/etc/kubernetes/pki/ca.crt``
* ``/etc/kubernetes/pki/ca.key``


## Further reading

See this [Vault article] for more details on how to integrate Vault as a
Kubernetes certificate manager.

<!--LINKS -->
[HashiCorp Vault]: https://developer.hashicorp.com/vault/docs
[bootstrap configuration file]: /snap/reference/config-files/bootstrap-config.md
[Vault article]: https://support.hashicorp.com/hc/en-us/articles/21920341210899-Create-an-Intermediate-CA-in-Kubernetes-using-Vault-as-a-certificate-manager
