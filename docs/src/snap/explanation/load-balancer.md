## Load balancer

{{ product }}' load balancer feature allows you to expose your workloads
externally. A load balancer is an important component that allows you to
distribute incoming network traffic from outside of your cluster to the
services inside.

## Kubernetes service types

Kubernetes allows you to expose your cluster's workloads in the following ways:

- **ClusterIP**: Exposes the service on a cluster-internal IP. With ClusterIP
  the service is only reachable from within the cluster.
- **NodePort**: Exposes the service on each Node's IP at a static port.
- **LoadBalancer**: Exposes a single IP address to distribute the incoming
  network traffic across multiple cluster nodes.

```{note}
The load balancer service type at layer 4 should not be confused with the Ingress Controller which operates at layer 7 (HTTP/HTTPS) and routes traffic from outside of your cluster to services inside of your cluster. Learn more about Ingress in the [Ingress guide][Ingress].
```

## Use case

The {{ product }}' load balancer feature is suitable for bare-metal setups and
private data-centers. If you are operating {{ product }} in a public cloud you
should consider using your cloud-provider's load balancing solution instead.

## IP address allocation

The load balancer feature assigns and unassigns IP addresses
for services of type LoadBalancer from a pool of IPs. Learn how to configure
the IP address pool with the [default load balancer] guide.

## Advertising IP address externally

The load balancer feature advertises the external IP address of the service
outside of the cluster. To accomplish this, the load balancer uses either Layer
2 mode or Border Gateway Protocol (BGP) Mode:

### Layer 2 Mode 

In Layer 2 mode, a cluster node takes the
responsibility of spreading the traffic for a service IP to its associated
pods using [kube-proxy]. Traffic routing relies on Address Resolution
Protocol ([ARP]) for IPv4 and Neighbor Discovery Protocol ([NDP])
for IPv6. While this mode is easy to implement, it is only suitable for
small-scale deployments. Consult the [layer 2 mode documentation] for more
information.

### BGP Mode

With BGP mode, neighboring routers exchange routing information
through peering sessions enabling efficient traffic distribution.
Traffic is balanced across nodes ensuring efficient resource utilization.
While this approach is more difficult to implement, it is suitable for
large-scale deployments. Refer to the [bgp mode documentation] for more
details.

## Traffic flow with a load balancer

When you expose a service with a load balancer, the following steps occur:

- A client sends a request to the external IP address allocated by the load
  balancer.
- The load balancer decides which Kubernetes node should handle the request
  based on the selected underlying mechanism (BGP or Layer 2).
- The node receiving the request routes it to a specific pod in the service
   with [kube-proxy]. 

## Next Steps

- Get started with configuring your cluster's [default load balancer].
- Create a service of the type LoadBalancer to expose your workloads externally
  by following the [upstream guide].
- Learn more about Ingress in the [Ingress guide][Ingress]

<!-- LINKS -->

[Ingress]: ingress
[default load balancer]: ../howto/networking/default-loadbalancer
[kube-proxy]: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-proxy/
[ARP]: https://en.wikipedia.org/wiki/Address_Resolution_Protocol
[NDP]: https://en.wikipedia.org/wiki/Neighbor_Discovery_Protocol
[layer 2 mode documentation]: https://metallb.io/concepts/layer2/
[bgp mode documentation]: https://metallb.io/concepts/bgp/
[upstream guide]: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#create-a-service
