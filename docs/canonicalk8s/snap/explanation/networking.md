# Networking

In Kubernetes, understanding how traffic is managed inside of your
cluster can be complex.
This page explains the networking features provided by {{product}} and
how they work together to handle traffic within and outside of your cluster.

## CNI (Container Network Interface)

Upstream Kubernetes does not include a built-in network provider and requires
a [network plugin] to handle pod-to-pod communication. {{product}} ships
with a default CNI that is automatically enabled when the cluster is
bootstrapped.

The CNI is responsible for:

- Assigning IP addresses to pods.
- Enabling communication between pods, even across different nodes.
- Managing network encapsulation and routing.

If you wish to use a different network plugin, please follow the
[alternative CNI] guide.

Learn how to manage the default network in the
[networking how-to guide][Network].

## Network

When you install {{product}}, the default network feature is automatically
enabled. The network feature configures the CNI to provide pod networking
across the cluster.

Key capabilities provided by the network feature:

- **Pod connectivity**: All pods can communicate with each other without NAT.
- **Service networking**: Kubernetes Services receive virtual IPs that route
  traffic to the correct set of pods.
- **Network policy**: Network policies can be used to restrict traffic between
  pods for security purposes. Learn more about [network policies][network
  policies].

## DNS

{{product}} includes a default DNS (Domain Name System) service which is
essential for internal cluster communication. When enabled, the DNS service
facilitates service discovery by assigning each Kubernetes Service a DNS
name. Pods can then reach services by name rather than by IP address.

Learn how to configure DNS in the [DNS how-to guide][DNS].

For more information on how DNS works in Kubernetes, consult the upstream
[DNS for Services and Pods][upstream DNS] documentation.

### Scaling DNS

Default DNS is scaled automatically through a Horizontal Pod Autoscaler
([HPA]) which monitors DNS pods' CPU and memory resource usage and adjusts
the number of replicas accordingly.

### DNS scheduling

DNS pods are scheduled with [topology spread constraints] to spread them
across the cluster nodes and zones when applicable. When {{product}} detects
that all pods are scheduled on the same node, it will restart the DNS pods
to re-balance their distribution.

A [priority class] is assigned to DNS pods to ensure their scheduling
before pods which are not node critical.

### DNS maintenance

When performing maintenance operations on the cluster, be aware that
the pod disruption budget ([PDB]) will only allow one DNS pod to be taken
down at a time.

## Load balancer

The load balancer feature allows you to expose your workloads externally
and distribute incoming network traffic from outside of your cluster to the
services inside. The load balancer feature is suitable for bare-metal setups
and private data centers. If you are operating {{product}} in a public cloud
you should evaluate your cloud provider's load balancing solution as well.

Learn how to configure the load balancer in the
[load balancer how-to guide][load balancer].

### IP address allocation

The load balancer feature assigns and releases IP addresses for services of
type `LoadBalancer` from a configured pool of IPs. Learn how to configure the
IP address pool with the [load balancer how-to guide][load balancer].

### Advertising IP addresses externally

The load balancer feature advertises the external IP address of the service
outside of the cluster. To accomplish this, the load balancer uses either
Layer 2 mode or Border Gateway Protocol (BGP) mode.

#### Layer 2 mode

In Layer 2 mode, a cluster node takes responsibility for advertising the
service IP to the local network. Traffic routing relies on Address Resolution
Protocol ([ARP]) for IPv4 and Neighbor Discovery Protocol ([NDP]) for IPv6.
While this mode is straightforward to set up, it is best suited for
small-scale deployments. Consult the [Layer 2 mode documentation] for more
information.

#### BGP mode

With BGP mode, cluster nodes establish peering sessions with neighboring
routers to exchange routing information, enabling efficient traffic
distribution. Traffic is balanced across nodes ensuring efficient resource
utilization. While this approach requires more configuration, it is suitable
for large-scale deployments. Refer to the [BGP mode documentation] for more
details.

### Traffic flow with a load balancer

When you expose a service with a load balancer, the following steps occur:

1. A client sends a request to the external IP address allocated by the load
   balancer.
2. The load balancer directs the request to a Kubernetes node based on the
   selected mode (BGP or Layer 2).
3. [kube-proxy] routes the request to a pod backing the service.

Create a service of type `LoadBalancer` to expose your workloads externally
by following the [upstream guide] or consult the
[load balancer how-to guide][load balancer].

## Ingress

[Ingress][Ingress K8s] is a Kubernetes resource that manages external HTTP
and HTTPS access to services within your cluster. Traffic routed through an
Ingress resource is directed to a service, which in turn forwards it to the
relevant pod running the desired application.

With {{product}}, Ingress is not enabled by default. Once enabled (see the
[Ingress how-to guide][Ingress]), you will have a working Ingress Controller
in your cluster.

Learn more about Ingress in the upstream
[Ingress documentation][Ingress K8s].

### Ingress controller

The Ingress resource defines the routing rules for incoming traffic, while the
[Ingress Controller] is responsible for implementing those rules by configuring
the underlying networking infrastructure of the cluster. Ingress does not work
without an Ingress Controller.

The Ingress Controller serves as a Layer 7 (HTTP/HTTPS) load balancer that
routes traffic from outside of your cluster to services inside of your cluster.
This should not be confused with the Kubernetes Service `LoadBalancer` type
which operates at Layer 4 and routes traffic directly to individual pods.

If your cluster requires different Ingress Controllers, the responsibility
of implementation falls upon you.

## Gateway

The [Gateway API][Gateway API] is a more expressive and extensible alternative
to Ingress for managing how traffic enters your cluster. When enabled,
{{product}} deploys the necessary Custom Resource Definitions (CRDs) and a
default `GatewayClass` to configure traffic routing and infrastructure
provisioning.

With {{product}}, Gateway is not enabled by default. Once enabled (see the
[Gateway how-to guide][Gateway]), you can define `Gateway` and `HTTPRoute`
resources to control how traffic is routed to your services.

Key capabilities of the Gateway feature:

- **GatewayClass**: A default `GatewayClass` named `ck-gateway` is provided
  when the feature is enabled.
- **Traffic routing**: Use `HTTPRoute` and other route types to define how
  requests are matched and forwarded to backend services.
- **Advanced traffic management**: The Gateway API supports traffic splitting,
  header-based routing, and other advanced patterns.

Learn more about Gateway API in the upstream
[Gateway API documentation][Gateway API].

<!-- LINKS -->
[PDB]: https://kubernetes.io/docs/tasks/run-application/configure-pdb/#specifying-a-poddisruptionbudget
[HPA]: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[topology spread constraints]: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
[priority class]: https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass

[load balancer]: /snap/howto/networking/default-loadbalancer
[kube-proxy]: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-proxy/
[ARP]: https://en.wikipedia.org/wiki/Address_Resolution_Protocol
[NDP]: https://en.wikipedia.org/wiki/Neighbor_Discovery_Protocol
[Layer 2 mode documentation]: https://metallb.io/concepts/layer2/
[BGP mode documentation]: https://metallb.io/concepts/bgp/
[upstream guide]: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#create-a-service

[alternative CNI]: /snap/howto/networking/alternative-cni
[Ingress]: /snap/howto/networking/default-ingress
[Network]: /snap/howto/networking/default-network
[DNS]: /snap/howto/networking/default-dns
[Gateway]: /snap/howto/networking/default-gateway
[network plugin]: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/
[network policies]: https://kubernetes.io/docs/concepts/services-networking/network-policies/
[upstream DNS]: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
[Ingress K8s]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
[Gateway API]: https://gateway-api.sigs.k8s.io/
