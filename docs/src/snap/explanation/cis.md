# CIS hardening

CIS Hardening refers to the process of implementing security configurations that
align with the benchmarks set forth by the [Center for Internet Security] (CIS).
These [benchmarks] are a set of best practices and guidelines designed to secure
various software and hardware systems, including Kubernetes clusters. The
primary goal of CIS hardening is to reduce the attack surface and enhance the
overall security posture of an environment by enforcing configurations that are
known to protect against common vulnerabilities and threats.

## Why is CIS hardening important for Kubernetes?

Kubernetes, by its nature, is a complex system with many components interacting
in a distributed environment. This complexity can introduce numerous security
risks if not properly managed such as unauthorised access, data breaches and
service disruption. CIS hardening for Kubernetes focuses on configuring various
components of a Kubernetes cluster to meet the security standards specified in
the [CIS Kubernetes Benchmark].

## Apply CIS hardening to {{product}}

If you would like to apply CIS hardening to your cluster see our [how-to guide].

<!-- LINKS -->
[benchmarks]: https://www.cisecurity.org/cis-benchmarks
[Center for Internet Security]: https://www.cisecurity.org/
[CIS Kubernetes Benchmark]: https://www.cisecurity.org/benchmark/kubernetes
[how-to guide]: ../howto/cis-hardening.md
