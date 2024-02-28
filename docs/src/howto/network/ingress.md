# How to use default Ingress

Canonical Kubernetes allows you to configure Ingress into your cluster. When enabled, it tells your cluster how external HTTP and HTTPS traffic should be routed to its services.

## What you'll need

This guide assumes the following:

- You are installing on Ubuntu 22.04 or later, **or** another OS which supports
  snap packages (see [snapd support])
- You have root or sudo access to the machine
- You have an internet connection
- The target machine has sufficient memory and disk space. To accommodate
  workloads, we recommend a system with ***at least*** 20G of disk space and 4G of
  memory.
- You have Canonical Kubernetes installed and a bootstraped cluster. (See: [getting-started-guide](#TODO))

## Is Ingress enabled?

Find out wether you have enabled Ingress with the following command:

```bash
sudo k8s status
```

The default state for the cluster is `ingress disabled`.

## Enabling and disabling Ingress
To enable Ingress, run:

```bash
sudo k8s enable ingress
```

To revoke this action use `disable`:

```bash
sudo k8s disable ingress
```

For more information on these two commands, execute:

```bash
sudo k8s help enable
```

Or for disabling:

```bash
sudo k8s help disable
```

To continue with the `Configuring Ingress` section enable ingress again.

## Configuring Ingress
Discover your configuration options by running:

```bash
sudo k8s set ingress –help
```

You should see three options:
- default-tls-secret: Name of the TLS (Transport Layer Security) Secret in the kube-system namespace 
that will be used as the default Ingress certificate
- enable-proxy-protocol: If set, proxy protocol will be enabled for the Ingress

### TLS Secret

You can create a tls secret by following the official kubernetes docs: [kubectl-create-secret-tls/](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_create/kubectl_create_secret_tls/).
Note: remember to use `sudo k8s kubectl` (See: [kubectl-guide](#TODO)).

Tell Ingress to use your new Ingress certificate:
```bash
sudo k8s set ingress --default-tls-secret=<new-default-tls-secret>
```

Replace `<new-default-tls-secret>` with the desired value for your Ingress configuration.

### Proxy Protocol
Enabling the proxy protocol allows passing client connection
 information to the backend service. 

Consult the official kubernetes documentation on [proxy-protocol](https://kubernetes.io/docs/reference/networking/service-protocols/#protocol-proxy-special).

Use the following command to enable the proxy protocol:

```bash
sudo k8s set ingress --enable-proxy-protocol=<new-enable-proxy-protocol>
```

Adjust the value of `<new-enable-proxy-protocol>` with your proxy protocol requirements.


<!-- LINKS -->

[Component Upgrades]: #TODO
[getting-started-guide]: (#TODO)

