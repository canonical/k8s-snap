# How to refresh externally managed Kubernetes certificates

This guide walks you through refreshing external certificates for both control
plane and worker nodes in your {{product}} cluster.

## Prerequisites

- A running {{product}} cluster

```{note} To refresh certificates, your cluster must have been initially
configured with external certificates during the bootstrap process. You can
verify which certificates are externally managed by running `k8s certs-status`
on a control plane node.
```

## Assemble certificates data

To simplify the process and avoid complex CLI commands, the `refresh-certs`
command accepts new node external certificates via the `--external-certificates`
argument with a YAML-formatted file. For a complete list of available
certificate keys, see the
[certificates refresh configuration file reference page][reference page].

If your cluster uses a mixed certificate management approach where
some certificates are managed externally and others internally, you must
explicitly specify the internally managed certificates to refresh on worker
nodes using the `--certificates` flag. Externally managed certificates should
continue to be provided through the `--external-certificates` argument.

Refer to the {{ product }}
[cluster certificates and configuration directories][certificates]
documentation to determine which
certificates are required for each node.

If you are managing some of the Certificate Authorities (CAs)
externally, provide only the certificates that require updates. Identify the
externally managed CAs by running `k8s certs-status` on a control plane node.

## Refresh Control Plane node certificates

Execute the following command to refresh certificates on each control plane
node:

```
sudo k8s refresh-certs --external-certificates ./certificates.yaml
```

If your node setup includes additional SANs, be sure to include the
specific SANs for each node when requesting new certificates from your
certificates authority. Check your provider's documentation for instructions on
requesting certificates with the required SANs.


The node will validate the certificates, update them automatically, and restart
the necessary services. Upon successful completion, you will see:

```
External certificates have been successfully refreshed.
```

Verify the new expiration dates by running the `k8s certs-status` command:

```
CERTIFICATE               EXPIRES                 RESIDUAL TIME  CERTIFICATE AUTHORITY      EXTERNALLY MANAGED
apiserver                 Mar 21, 2026 01:06 UTC  364d           kubernetes-ca              yes
apiserver-kubelet-client  Mar 21, 2026 01:06 UTC  364d           kubernetes-ca-client       yes
front-proxy-client        Mar 21, 2026 01:06 UTC  364d           kubernetes-front-proxy-ca  yes
kubelet                   Mar 21, 2026 01:06 UTC  364d           kubernetes-ca              yes
admin.conf                Mar 21, 2026 01:06 UTC  364d           kubernetes-ca-client       yes
controller.conf           Mar 21, 2026 01:06 UTC  364d           kubernetes-ca-client       yes
kubelet.conf              Mar 21, 2026 01:06 UTC  364d           kubernetes-ca-client       yes
proxy.conf                Mar 21, 2026 01:06 UTC  364d           kubernetes-ca-client       yes
scheduler.conf            Mar 21, 2026 01:06 UTC  364d           kubernetes-ca-client       yes

CERTIFICATE AUTHORITY      EXPIRES                 RESIDUAL TIME  EXTERNALLY MANAGED
kubernetes-ca              Mar 19, 2035 01:06 UTC  9y             yes
kubernetes-ca-client       Mar 19, 2035 01:06 UTC  9y             yes
kubernetes-front-proxy-ca  Mar 19, 2035 01:06 UTC  9y             yes
```

## Refresh Worker node certificates

To refresh the certificates on worker nodes, perform this step on each worker
node in your cluster:

```
sudo k8s refresh-certs --external-certificates ./certificates.yaml
```

The node will automatically update certificates and restart necessary services.
Upon successful completion, you will see:

```
External certificates have been successfully refreshed.
```

<!-- Links -->

[certificates]: /snap/reference/certificates.md
[reference page]: /snap/reference/config-files/refresh-external-certs-config.md
