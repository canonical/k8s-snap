# Networking 

In Kubernetes, understanding how traffic is managed inside of your
cluster can be complex.
This explanation provides you with the a more in-depth explanation of
how networking is handled in {{product}}.

## Network

When you install {{product}}, the default Network is automatically enabled.

Since upstream Kubernetes comes without a network provider,
it requires the use of a [network plugin][network plugin].
This network plugin facilitates communication between pods,
services, and external resources, ensuring smooth traffic flow within the
cluster. The current implementation of {{product}} leverages a widely adopted
CNI (Container Network Interface) called [Cilium][Cilium].
If you wish to use a different network plugin
please follow the [alternative CNI] guide.

Learn how to use the {{product}} default network
in the [networking how-to guide][Network].

## Load balancer

{{ product }}' load balancer feature allows you to expose your workloads
externally and distribute incoming network traffic from outside of your 
cluster to the services inside. The {{ product }}' load balancer feature is 
suitable for bare-metal setups and private data-centers. If you are operating 
{{ product }} in a public cloud you should evaluate your cloud-provider's load 
balancing solution as well.

### IP address allocation

The load balancer feature assigns and releases IP addresses
for services of type LoadBalancer from a pool of IPs. Learn how to configure
the IP address pool with the [default load balancer] guide.

### Advertising IP address externally

The load balancer feature advertises the external IP address of the service
outside of the cluster. To accomplish this, the load balancer uses either Layer
2 mode or Border Gateway Protocol (BGP) Mode.

#### Layer 2 Mode 

In Layer 2 mode, a cluster node takes the
responsibility of spreading the traffic for a service IP to its associated
pods using [kube-proxy]. Traffic routing relies on Address Resolution
Protocol ([ARP]) for IPv4 and Neighbor Discovery Protocol ([NDP])
for IPv6. While this mode is easy to implement, it is only suitable for
small-scale deployments. Consult the [layer 2 mode documentation] for more
information.

#### BGP Mode

With BGP mode, neighboring routers exchange routing information
through peering sessions enabling efficient traffic distribution.
Traffic is balanced across nodes ensuring efficient resource utilization.
While this approach is more difficult to implement, it is suitable for
large-scale deployments. Refer to the [bgp mode documentation] for more
details.

### Traffic flow with a load balancer

When you expose a service with a load balancer, the following steps occur:

- A client sends a request to the external IP address allocated by the load
  balancer.
- The load balancer decides which Kubernetes node should handle the request
  based on the selected underlying mechanism (BGP or Layer 2).
- kube-proxy routes the request to a specific pod in the service. 

Create a service of the type LoadBalancer to expose your workloads externally
by following the [upstream guide] or consult our [default load balancer] guide.

## Ingress

[Ingress][Ingress K8s] is a Kubernetes resource that manages
external access by handling both HTTP and HTTPS traffic to services within
your cluster. Traffic routed through the Ingress is directed to a service,
which in turn forwards it to the relevant pod running the desired application 
within a container.

The underlying mechanism provided by default is currently Cilium.
However, it should always be operated through the provided CLI rather than
directly. This way, we can provide the best experience for future cluster
maintenance and upgrades.

With {{product}}, Ingress is not enabled by default. Once enabled 
(see the [default Ingress guide][Ingress]), you will have a working
Ingress Controller in your cluster. 

### Ingress controller

The Ingress resource defines the routing rules for the incoming traffic while 
the [Ingress Controller][Ingress Controller] is responsible for implementing
those rules by configuring the underlying networking infrastructure of
the cluster. Ingress does not work without an Ingress Controller.

The Ingress Controller also serves as a layer 7 (HTTP/HTTPS) load balancer
that routes traffic from outside of your cluster to services
inside of your cluster. Please do not confuse this with the
Kubernetes Service LoadBalancer type which operates at layer 4 and routes
traffic directly to individual pods.

![cluster6][]

If your cluster requires different Ingress Controllers,
the responsibility of implementation falls upon you. 

## DNS

{{product}} includes a default DNS (Domain Name System) solution which is
essential for internal cluster communication. When enabled, the DNS facilitates
service discovery by assigning each service a DNS name.

### Scaling DNS

Default DNS is scaled automatically through a Horizontal Pod Autoscaler ([HPA])
which monitors DNS pods' CPU and memory resource usage and adjusts the number of
replicas accordingly.

### DNS scheduling

DNS pods are scheduled with [topology spread constraints] to spread them across
the cluster nodes and zones when applicable.
When {{ product }} detects that all pods are scheduled on the same node, it
will restart the DNS pods to re-balance their distribution.

A [priority class] is assigned to DNS pods to ensure their scheduling
before pods which are not node critical.

### DNS maintenance

When performing maintenance operations on the cluster, please be aware that
the pod disruption budget ([PDB]) will only allow one DNS pod to be taken down
at a time.

<!-- IMAGES -->

[cluster6]: https://assets.ubuntu.com/v1/e6d02e9c-cluster6.svg

<!-- LINKS -->
[PDB]: https://kubernetes.io/docs/tasks/run-application/configure-pdb/#specifying-a-poddisruptionbudget
[HPA]: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[topology spread constraints]: https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/
[priority class]: https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass

[default load balancer]: /snap/howto/networking/default-loadbalancer
[kube-proxy]: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-proxy/
[ARP]: https://en.wikipedia.org/wiki/Address_Resolution_Protocol
[NDP]: https://en.wikipedia.org/wiki/Neighbor_Discovery_Protocol
[layer 2 mode documentation]: https://metallb.io/concepts/layer2/
[bgp mode documentation]: https://metallb.io/concepts/bgp/
[upstream guide]: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#create-a-service

[alternative CNI]: /snap/howto/networking/alternative-cni
[Ingress]: /snap/howto/networking/default-ingress
[Network]: /snap/howto/networking/default-network
[Cilium]: https://cilium.io/
[network plugin]: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/
[Ingress K8s]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
