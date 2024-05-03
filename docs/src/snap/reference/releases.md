# Release notes

## Rolling preview release

In advance of a GA release of Canonical Kubernetes, you can still install and
try out the newest distribution of Kubernetes.

You need two commands to get a single node cluster, one for installation and
another for cluster bootstrap. You can try it out now on your console by
installing the k8s snap from the beta channel:

```
sudo snap install k8s --channel=1.30-classic/beta --classic
sudo k8s bootstrap
```

Currently Canonical Kubernetes is working towards general availability, but you
can install it now to try:

- **Clustering** - need high availability or just an army of worker nodes?
  Canonical Kubernetes is emminently scaleable, see the [tutorial on adding
  more nodes][nodes]. 
- **Networking** - Our built-in network component allows cluster administrators
  to automatically scale and secure network policies across the cluster. Find
  out more in our [how-to guide][networking].
- **Observability** - Canonical Kubernetes ships with [COS Lite], so you never
  need to wonder what your cluster is actually doing. See the [observability
  documentation] for more details.

Follow along with the [tutorial] to get started!


<!-- LINKS -->

[tutorial]: /snap/tutorial/getting-started
[nodes]: /snap/tutorial/add-remove-nodes
[COS Lite]: https://charmhub.io/cos-lite
[networking]: /snap/howto/networking
[observability documentation]: /snap/howto/cos-lite