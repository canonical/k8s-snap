# Ingress

In Kubernetes, understanding how inbound traffic is managed inside of your
cluster can be complex.
This explanation provides you with the essentials
to successfully manage your {{product}} cluster.

## Network

When you install {{product}}, the default Network is automatically enabled.
This is also a requirement for the default Ingress to function.

Since upstream Kubernetes comes without a network provider,
it requires the use of a [network plugin][network plugin].
This network plugin facilitates communication between pods,
services, and external resources, ensuring smooth traffic flow within the cluster.
The current implementation of {{product}} leverages a widely adopted
CNI (Container Network Interface) called [Cilium][Cilium].
If you wish to use a different network plugin
please follow the [alternative-cni] guide.

Learn how to use the {{product}} default network in the [networking how-to guide][Network].

## Kubernetes Pods and Services

In Kubernetes, the smallest unit is a pod, which encapsulates application containers.
Since pods are ephemeral and their IP addresses change when destroyed and restarted,
they are exposed through services.
Services offer a stable network interface by providing discoverable names and
load balancing functionality for managing a set of pods.
For further details on Kubernetes Services,
refer to the [upstream Kubernetes Service documentation][Service].

## Ingress

[Ingress][Ingress K8s] is a Kubernetes resource that manages
external access by handling both HTTP and HTTPS traffic to services within your cluster.
Traffic routed through the Ingress is directed to a service,
which in turn forwards it to the relevant pod
running the desired application within a container.

The Ingress resource lets you define rules on how traffic should get handled.
Refer to the [Kubernetes documentation on Ingress rules][Ingress Rules]
for up to date information on the available rules and their implementation.


While the Ingress resource manages the routing rules for the incoming traffic,
the [Ingress Controller][Ingress Controller] is responsible for implementing
those rules by configuring the underlying networking infrastructure of the cluster.
Ingress does not work without an Ingress Controller.

The Ingress Controller also serves as a layer 7 (HTTP/HTTPS) load balancer
that routes traffic from outside of your cluster to services inside of your cluster.
Please do not confuse this with the Kubernetes Service LoadBalancer type
which operates at layer 4 and routes traffic directly to individual pods.

![cluster6][]

With {{product}}, enabling Ingress is easy:
See the [default Ingress guide][Ingress].
Once enabled, you will have a working
[Ingress Controller][Cilium Ingress Controller] in your cluster.

The underlying mechanism provided by default is currently Cilium.
However, it should always be operated through the provided CLI rather than
directly. This way, we can provide the best experience for future cluster
maintenance and upgrades.

If your cluster requires different Ingress Controllers,
the responsibility of implementation falls upon you.

You will need to create the Ingress resource,
outlining rules that direct traffic to your application's Kubernetes service.

<!-- IMAGES -->

[cluster6]: https://assets.ubuntu.com/v1/e6d02e9c-cluster6.svg

<!-- LINKS -->

[alternative-cni]: ../howto/networking/alternative-cni
[Ingress]: ../howto/networking/default-ingress
[Network]: ../howto/networking/default-network
[LoadBalancer]: load-balancer
[Cilium]: https://cilium.io/
[network plugin]: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/
[Service]: https://kubernetes.io/docs/concepts/services-networking/service/
[Ingress K8s]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[Ingress Rules]: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules
[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
[Cilium Ingress Controller]: https://docs.cilium.io/en/stable/network/servicemesh/ingress/
