# How to use default Ingress

<!-- SPREAD SUITE: snap_bootstrapped -->

{{product}} enables you to configure Ingress for your cluster. When enabled, it
directs external HTTP and HTTPS traffic to the appropriate services within the
cluster.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstrapped {{product}} cluster (see the [Getting
   Started](/snap/tutorial/getting-started) guide).

## Check Ingress status

Find out whether Ingress is enabled or disabled with the following command:

```
sudo k8s status
```

<!-- SPREAD
sudo k8s get ingress | grep "enabled: false"
-->

Please ensure that Ingress is enabled on your cluster.

## Enable Ingress

To enable Ingress, run:

```
sudo k8s enable ingress
```

<!-- SPREAD
sudo k8s get ingress | grep "enabled: true"
--> 

For more information on the command, execute:

```
sudo k8s help enable
```

<!-- SPREAD
sudo k8s help enable | grep "Enable one of network, dns"
-->

```{warning}
The Kubernetes Service created for the ingress controller is set to single
stack networking(IPv4 or IPv6) by default. The Service can be manually
patched with `ipFamilyPolicy: PreferDualStack` to enable dual stack networking.
```

## Configure Ingress

Discover your configuration options by running:

```
sudo k8s get ingress
```

<!-- SPREAD
sudo k8s get ingress | grep "enabled: true"
-->

You should see three options:

- `enabled`: If set to true, Ingress is enabled
- `default-tls-secret`: Name of the TLS (Transport Layer Security) Secret in
   the kube-system namespace that will be used as the default Ingress
   certificate
- `enable-proxy-protocol`: If set, proxy protocol will be enabled for the
   Ingress

### TLS secret

You can create a TLS secret by following the official
[Kubernetes documentation](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_create/kubectl_create_secret_tls/).
Please remember to use `sudo k8s kubectl` (See the [kubectl-guide](/snap/tutorial/kubectl)).

Tell Ingress to use your new Ingress certificate:

<!-- SPREAD SKIP -->

```
sudo k8s set ingress.default-tls-secret=<new-default-tls-secret>
```

Replace `<new-default-tls-secret>` with the desired value for your Ingress
configuration.

<!-- SPREAD SKIP END -->

<!-- SPREAD 
sudo k8s set ingress.default-tls-secret=new-default-tls-secret
sudo k8s get ingress | grep "default-tls-secret: new-default-tls-secret"
-->

### Proxy protocol

Enabling the proxy protocol allows passing client connection information to the
backend service.

Consult the official
[Kubernetes documentation on the proxy protocol](https://kubernetes.io/docs/reference/networking/service-protocols/#protocol-proxy-special).

Use the following command to enable the proxy protocol:

```
sudo k8s set ingress.enable-proxy-protocol=true
```

<!-- SPREAD 
sudo k8s get ingress | grep "enable-proxy-protocol: true"
-->

## Disable Ingress

You can `disable` the built-in ingress:

```{warning}
Disabling Ingress may impact external access to services within your cluster.
Ensure that you have alternative configurations in place before disabling Ingress.
```

```
sudo k8s disable ingress
```

<!-- SPREAD
sudo k8s get ingress | grep "enabled: false"
-->

For more information on this command, run:

```
sudo k8s help disable
```

<!-- SPREAD
sudo k8s help disable | grep "Disable one of network, dns"
--> 

## Next Step

- Learn more in the [networking explanation](/snap/explanation/networking.md#ingress) page.

<!-- LINKS -->

[kubectl-create-secret-TLS/]: https://kubernetes.io/docs/reference/kubectl/generated/kubectl_create/kubectl_create_secret_tls/
[proxy-protocol]: https://kubernetes.io/docs/reference/networking/service-protocols/#protocol-proxy-special
[getting-started-guide]: /snap/tutorial/getting-started
[kubectl-guide]: /snap/tutorial/kubectl
