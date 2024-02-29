# How to use default Ingress

Canonical Kubernetes allows you to configure Ingress into your cluster.
When enabled, it tells your cluster how external HTTP and HTTPS traffic should be routed to its services.

## What you'll need

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstraped Canonical Kubernetes cluster (see the [Getting Started][getting-started-guide] guide).

## Check Ingress status

Find out whether Ingress is enabled or disabled with the following command:

```bash
sudo k8s status
```

The default state for the cluster is `ingress disabled`.

## Enable Ingress

To enable Ingress, run:

```bash
sudo k8s enable ingress
```

For more information on the command, execute:

```bash
sudo k8s help enable
```

## Configure Ingress

Discover your configuration options by running:

```bash
sudo k8s set ingress –help
```

You should see three options:

- `default-tls-secret`: Name of the TLS (Transport Layer Security) Secret in the kube-system namespace 
  that will be used as the default Ingress certificate
- `enable-proxy-protocol`: If set, proxy protocol will be enabled for the Ingress

### TLS Secret

You can create a TLS secret by following the official [Kubernetes documentation][kubectl-create-secret-tls/].
Note: remember to use `sudo k8s kubectl` (See the [kubectl-guide]).

Tell Ingress to use your new Ingress certificate:
```bash
sudo k8s set ingress.default-tls-secret=<new-default-tls-secret>
```

Replace `<new-default-tls-secret>` with the desired value for your Ingress configuration.

### Proxy Protocol

Enabling the proxy protocol allows passing client connection information to the backend service. 

Consult the official [Kubernetes documentation on the proxy protocol][proxy-protocol].

Use the following command to enable the proxy protocol:

```bash
sudo k8s set ingress.enable-proxy-protocol=<new-enable-proxy-protocol>
```

Adjust the value of `<new-enable-proxy-protocol>` with your proxy protocol requirements.

## Disable Ingress

You can `disable` the built-in ingress:

``` {warning} Disabling Ingress may impact external access to services within your cluster.
    Ensure that you have alternative configurations in place before disabling Ingress.
```

```bash
sudo k8s disable ingress
```

For more information on this command, run:

```bash
sudo k8s help disable
```

<!-- LINKS -->

[kubectl-create-secret-tls/]: https://kubernetes.io/docs/reference/kubectl/generated/kubectl_create/kubectl_create_secret_tls/
[proxy-protocol]: https://kubernetes.io/docs/reference/networking/service-protocols/#protocol-proxy-special
[getting-started-guide]: ../../../tutorial/getting-started
[kubectl-guide]: ../../../tutorial/kubectl
