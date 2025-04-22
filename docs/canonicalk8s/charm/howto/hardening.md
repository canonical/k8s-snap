# How to harden your {{product}} cluster

The {{product}} hardening guide provides actionable steps to enhance the
security posture of your deployment. These steps are designed to help you align
with industry-standard frameworks such as CIS and DISA STIG.

{{product}} aligns with many security recommendations by
default. However, since implementing all security recommendations
would come at the expense of compatibility and/or performance we expect
cluster administrators to follow post deployment hardening steps based on their
needs.

This how-to has both the recommended minimum hardening steps and also a more
comprehensive list of manual tests.

Please evaluate the implications of each configuration before applying it.

## Platform hardening recommendations

These steps are common to the hardening process for not only CIS and DISA STIG
compliance, but also good suggestions if one is willing to incur the performance
costs for the benefit of an increased security posture.

```{note}
The following guide defines various service arguments by modifying the
`/var/snap/k8s/common/args/<service>` files.

When using the charm, you may either connect to the Juju units and perform
the same steps manually *or* use charm settings to specify the list of
service arguments, for example through `kube-apiserver-extra-args`.
```

```{include} /_parts/common_hardening.md
```


## CIS and DISA STIG hardening

To assess compliance to DISA STIG recommendations, please see
[DISA STIG assessment page].

To assess compliance to the CIS hardening guidelines, please see the [CIS
assessment page].

<!-- Links -->
[DISA STIG assessment page]: /snap/how-to/security/disa-stig-assessment.md
[CIS assessment page]: /snap/how-to/security/cis-assessment.md
