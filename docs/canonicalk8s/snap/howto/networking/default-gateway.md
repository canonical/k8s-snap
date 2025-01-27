# How to use default Gateway
<!-- does gateway get a capital letter - like ingress -->
<!-- what does gateway do -->
{{product}} enables you to configure networking of your cluster using
[gateway API].

<!-- enabled by default ? -->

 Ingress for your cluster. When enabled, it
directs external HTTP and HTTPS traffic to the appropriate services within the
cluster.

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine
- You have a bootstrapped {{product}} cluster (see the [Getting
  Started][getting-started-guide] guide).

## Check gateway status


Find out whether gateway is enabled or disabled with the following command:

```
sudo k8s status
```

Please ensure that gateway is enabled on your cluster.

## Enable gateway

To enable gateway, run:

```
sudo k8s enable gateway
```

###
<!-- is there anything else needed with gateway config -->


## Disable gateway

You can `disable` the built-in gateway:
<!-- are there any warnings for gateway -->
```{warning}
Disabling Ingress may impact external access to services within your cluster.
Ensure that you have alternative configurations in place before disabling Ingress.
```

```
sudo k8s disable gateway
```

<!-- LINKS -->

[gateway API]:https://gateway-api.sigs.k8s.io/
[getting-started-guide]: ../../tutorial/getting-started
[kubectl-guide]: ../../tutorial/kubectl
