# Ingress
In Kubernetes, understanding how inbound traffic is managed inside of your cluster can be complex.
While there is plenty of upstream documentation on the topic, 
this guide is tailored towards what you need to know to successfully manage your Canonical Kubernetes cluster.

## What is Ingress?
[Ingress][Ingress] is a Kubernetes Resource that is responsible for managing external access 
(HTTP and HTTPS traffic) to services within your cluster. 

The Ingress object allows you to define rules on how traffic should get handled. See [Ingress Rules][Ingress Rules].

While the Ingress Resource manages the routing rules for the incoming traffic, 
the [Ingress Controller][Ingress Controller] is responsible for implementing those rules 
by configuring the underlying networking infrastructure of the cluster.
Ingress does not work without an Ingress Controller. This controller evaluates the rules that are defined in the Ingress Resource to determine traffic routing.

With Canonical Kubernetes, enabling default ingress is easy: [Ingress][Ingress]



## Kubernetes Services
When the Ingress provider forwards the traffic to a service, 
the service forwards the traffic to a pod on which a desired application may be running inside of a container.
Services provide discoverable names and load balancing, for when it is dealing with a set of pods.
As pods get destroyed and restarted their IP address changes 
which is why services are used to provide us with a stable network interface.

Learn more about Kubernetes Services in the upstream docs: [Kubernetes Service][Service].

## Network
You may only use Canonical Kubernetes default Ingress when the Canonical Kubernetes default network is enabled on your cluster. 

Since upstream Kubernetes comes without a network provider, it requires the use of a [network plugin][network plugin].
At the moment of writing, Canonical Kubernetes leverages a popular CNI (Container Network Interface) called [Cilium][Cilium]. 


With Canonical Kubernetes, a network plugin is enabled by default.
Learn more about our default network: [Network][Network]


## Kubernetes Networking: Who needs to talk to who?
In th Kubernetes Network Model there are four types of communication:

1. Container-to-Container communication
2. Pod-to-Pod communication
3. Pod-to-Service communication
4. Node-to-node communication

In Kubernetes we prefer the use of services to expose our applications. 
External traffic is routed to a kubernetes service by our Ingress Controller 
which forwards it to the pod on which our application is running.
We make use of services as pods IP are not stable over time.


<!-- LINKS -->

[Ingress]: /snap/howto/networking/default-ingress
[Network]: /snap/howto/networking/default-network
[Cilium]: https://cilium.io/
[network plugin]: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/
[Service]: https://kubernetes.io/docs/concepts/services-networking/service/
[Ingress]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[Ingress Rules]: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules
[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/