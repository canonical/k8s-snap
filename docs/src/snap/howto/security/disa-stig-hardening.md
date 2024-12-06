# DISA STIG for {{product}}

Security Technical Implementation Guides (STIG) are developed by the Defense
Information System Agency (DISA) for the U.S. Department of Defense (DoD).

The Kubernetes STIGs contain guidelines on how to check remediate various
potential security concerns for a Kubernetes deployment.

This document lists the steps a Kubernetes system administrator or auditor must
take verify each STIG Finding against the k8s-snap.


## What you'll need

This guide assumes the following:

- You have a bootstrapped {{product}} cluster (see the [getting started] guide)
- You have root or sudo access to the machine


## Post-deployment configuration steps

{{product}} complies with most DISA STIG recommendations by default. However,
some checks require administrator consideration and intervention. You can
review these steps in the [Post-Deployment Configuration Steps][] section.

```{include} ../../_parts/common_hardening.md
```

## Additional DISA-STIG specific steps

TODO

## Manually audit DISA STIG hardening recommendations

For manual audits of CIS hardening recommendations, please visit the
[Comprehensive Hardening Checklist][].


<!-- Links -->
[Hardening]:security/hardening.md
[Center for Internet Security (CIS)]:https://www.cisecurity.org/
[kube-bench]:https://aquasecurity.github.io/kube-bench/v0.6.15/
[CIS Kubernetes Benchmark]:https://www.cisecurity.org/benchmark/kubernetes
[getting started]: ../tutorial/getting-started
[kube-bench release]: https://github.com/aquasecurity/kube-bench/releases
[Post-Deployment Configuration Steps]: security/hardening.md#post-deployment-configuration-steps
[Comprehensive Hardening Checklist]: security/hardening.md#comprehensive-hardening-checklist
