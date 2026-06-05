# How to expose the Kubernetes API server externally

By default, the {{product}} API server is reachable on each control plane
node's own address. External clients (`kubectl`, CI systems, identity
providers, etc.) therefore have to target one specific control plane IP and
lose access if that node goes down.

This guide shows how to put all of your API servers behind a single, stable
external IP using the built-in load balancer. Clients connect to that one
address and traffic is distributed across every healthy API server.

```{note}
This is a manual procedure built entirely from standard Kubernetes resources
and the existing `k8s` commands. The load balancer only forwards Layer 4 (TCP)
traffic, so TLS and client authentication stay end-to-end between the client
and the API servers.
```

## Prerequisites

This guide assumes the following:

- You have root or sudo access to the machine.
- You have a bootstrapped {{product}} cluster (see the [Getting
  Started][getting-started-guide] guide).
- Your cluster has self-signed certificates enabled (the default), so that
  certificates can be refreshed (see [How to refresh Kubernetes
  certificates][refresh-certs]).

## Enable and configure the load balancer

Enable the load balancer and configure an address pool. The external IP you
pick later must fall within one of these CIDRs:

```
sudo k8s set load-balancer.cidrs=10.0.10.0/24
sudo k8s set load-balancer.l2-mode=true
sudo k8s enable load-balancer
```

See [How to use the default load balancer][load-balancer] for the full set of
options (including BGP mode). Confirm the load balancer is running:

```
sudo k8s status
```

## Choose an external IP

Pick a free address from the configured pool. This is the single endpoint your
external clients will use. In this guide we use `10.0.10.50`.

## Create the external service and endpoints

Create a **selectorless** `LoadBalancer` service together with an
`EndpointSlice` that points at the `ClusterIP` of the `kubernetes` service in default.

```
sudo k8s get service kubernetes -owide

NAME         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   10.152.183.1   <none>        443/TCP   27h
```

Because the service has no selector, Kubernetes will not manage its endpoints
for you; the manually created `EndpointSlice` provides the backends, and
`kube-proxy` load-balances incoming traffic across them.

Save the following as `external-apiserver.yaml`, replacing the IP address
with the one from your cluster:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: kubernetes-external
  namespace: default
spec:
  type: LoadBalancer
  selector: {}                      # selectorless — backends come from the EndpointSlice
  externalTrafficPolicy: Local      # preserves client source IP (apiserver audit logs)
  loadBalancerIP: 10.0.10.50
  ports:
    - name: https
      port: 6443
      targetPort: 443
      protocol: TCP
---
apiVersion: discovery.k8s.io/v1
kind: EndpointSlice
metadata:
  name: kubernetes-external
  namespace: default
  labels:
    # Associates this EndpointSlice with the service above.
    kubernetes.io/service-name: kubernetes-external
addressType: IPv4
ports:
  - name: https
    port: 443
    protocol: TCP
endpoints:
  - addresses:
      - 10.152.183.1
```

Apply the manifest:

```
sudo k8s kubectl apply -f external-apiserver.yaml
```

Confirm the service was assigned the external IP you requested:

```
sudo k8s kubectl get service kubernetes-external -n default
```

```
NAME                  TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
kubernetes-external   LoadBalancer   10.152.183.200   10.0.10.50    6443:31234/TCP   10s
```

## Refresh the API server certificates

The external IP must appear in the API server serving certificate, otherwise
clients fail with `x509: certificate is valid for ..., not 10.0.10.50`.

Run the following on **each control plane node**, adding the external IP as an
extra SAN:

```
sudo k8s refresh-certs --extra-sans 10.0.10.50 --expires-in 1y
```

```{note}
Include any extra SANs your node already uses in the same command — the refresh
applies the SANs you pass, so omitting a previously added SAN drops it. See
[How to refresh Kubernetes certificates][refresh-certs] for details.
```

## Generate a kubeconfig for the external IP

Generate an admin kubeconfig that targets the external endpoint instead of the
local node address:

```
sudo k8s config --server https://10.0.10.50:6443 > external.kubeconfig
```

## Access the cluster

Copy `external.kubeconfig` to any machine that can reach the external IP and
use it:

```
KUBECONFIG=./external.kubeconfig kubectl get nodes
```

Requests now reach the external IP, are load-balanced across all healthy API
servers, and survive the loss of any single control plane node.

<!-- LINKS -->
[getting-started-guide]: ../tutorial/getting-started
[load-balancer]: networking/default-loadbalancer.md
[refresh-certs]: refresh-certs.md
[Subject Alternative Name]: https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.6
