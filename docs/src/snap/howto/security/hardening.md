# {{product}} hardening guide

The {{product}} hardening guide provides actionable steps to enhance the
security posture of your deployment. These steps are designed to help you align
with industry-standard frameworks such as CIS and DISA STIG.

{{product}} aligns with many security recommendations by
default. However, since implementing all security recommendations
would comes at the expense of compatibility and/or performance we expect
cluster administrators to follow post deployment hardening steps based on their
needs.

This how-to has both the recommended minimum hardening steps and also a more
comprehensive list of manual tests.

Please evaluate the implications of each configuration before applying it.

## Post-deployment hardening steps

These steps are common to the hardening process for both CIS and DISA STIG
compliance.

```{include} ../../../_parts/common_hardening.md
```


## CIS and DISA STIG Hardening

To assess compliance to DISA STIG recommendations, please see
[DISA STIG assessment page].

To assess compliance to the CIS hardening guidelines, please see the [CIS
assessment page].

<!-- Links -->
[upstream instructions]:https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/
[rate limits]:https://kubernetes.io/docs/reference/config-api/apiserver-eventratelimit.v1alpha1
[DISA STIG assessment page]: disa-stig-assessment.md
[CIS assessment page]: cis-assessment.md
