# Ingress

In Kubernetes, understanding how inbound traffic is managed inside of your
cluster can be complex.
While there is abundant documentation, this explanation provides you with the essentials
to successfully manage your Canonical Kubernetes cluster.

## Kubernetes Pods and Services

In Kubernetes, the smallest unit is a pod, which encapsulates application containers.
Since pods are ephemeral and their IP addresses change when destroyed and restarted,
they are exposed through services.
Services offer a stable network interface by providing discoverable names and
load balancing functionality for managing a set of pods.
For further details on Kubernetes Services,
refer to the upstream documentation:[Kubernetes Service][Service].

## Ingress

```{kroki} ../../assets/ingress.puml
```

[Ingress][Ingress K8s] is a Kubernetes Resource that manages
external access by handling both HTTP and HTTPS traffic to services within your cluster.
Traffic routed through the Ingress is directed to a service,
which in turn forwards it to the relevant pod
running the desired application within a container.

The Ingress Resource lets you define rules on how traffic should get handled.
Learn more about these rules and how to implement them here:
[Ingress Rules][Ingress Rules].

While the Ingress Resource manages the routing rules for the incoming traffic,
the [Ingress Controller][Ingress Controller] is responsible for implementing
those rules by configuring the underlying networking infrastructure of the cluster.
Ingress does not work without an Ingress Controller.

The Ingress Controller also serves as a layer 7 (HTTP/HTTPS) load balancer
that routes traffic from outside of your cluster to services inside of your cluster.
Please do not confuse this with the Kubernetes Service LoadBalancer type.

With Canonical Kubernetes, enabling default Ingress is easy:
See [default Ingress][Ingress].
At the moment of writing, enabling default Ingress will create a
[Cilium Ingress Controller][Cilium Ingress Controller] for you.
If your cluster requires different Ingress Controllers,
the responsibility of implementation falls upon you.

You will need to create the Ingress Resource,
outlining rules that direct traffic to your application's Kubernetes service.

## Network

In order to use Canonical Kubernetes default Ingress ensure that the
Canonical Kubernetes default network is enabled on your cluster.
This is the case by default.

Since upstream Kubernetes comes without a network provider,
it requires the use of a [network plugin][network plugin].
This network plugin facilitates communication between pods,
services, and external resources, ensuring smooth traffic flow within the cluster.
The current implementation of Canonical Kubernetes leverages a widely adopted
CNI (Container Network Interface) called [Cilium][Cilium].
If you wish to use a different network plugin
the implementation and configuration falls under your responsibility.

Learn how to use the Canonical Kubernetes default network: [Network][Network]

<!-- LINKS -->

[Ingress]: /snap/howto/networking/default-ingress
[Network]: /snap/howto/networking/default-network
[Cilium]: https://cilium.io/
[network plugin]: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/
[Service]: https://kubernetes.io/docs/concepts/services-networking/service/
[Ingress K8s]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[Ingress Rules]: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules
[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
[Cilium Ingress Controller]: https://docs.cilium.io/en/stable/network/servicemesh/ingress/
