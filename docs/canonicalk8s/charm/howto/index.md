# How-to guides

If you have a specific goal, and are already familiar with Kubernetes, our
How-to guides are more specific and contain less background information.
They’ll help you achieve an end result but may require you to understand and
adapt the steps to fit your specific requirements.

```{toctree}
:hidden:
Overview <self>
```

## Install and configure

Installation follows a similar pattern on all platforms, but some
differences must be addressed in each case. You may also want to customize the
installation of your Canonical Kubernetes nodes.

```{toctree}
:glob:
:titlesonly:
Install <install/index>
Configure the cluster <configure-cluster>
```

## Integrate

Learn how to integrate {{product}} with other charms to truly customize your
cluster and provide additional functionality.

```{toctree}
:glob:
:titlesonly:
Integrate with OpenStack <openstack>
Integrate with etcd <etcd>
Integrate with ceph-csi <ceph-csi>
```

## Networking

```{toctree}
:glob:
:titlesonly:
Configure proxy settings <proxy>
```

## Upgrade

Perform major and minor upgrades of your {{product}} cluster.

```{toctree}
:glob:
:titlesonly:
Upgrade minor version <upgrade-minor>
Upgrade patch version <upgrade-patch>

```

## Image registry

```{toctree}
:glob:
:titlesonly:
Configure a custom registry <custom-registry>
```

## Monitoring and troubleshooting

Sometimes things go wrong and you need to troubleshoot. Having observability
set up on your cluster can greatly increase the rate at which a problem is
identified and solved.

```{toctree}
:titlesonly:
Troubleshoot <troubleshooting>
Validate the cluster <validate>
Set up cluster observability <cos-lite>
```

## Security

```{toctree}
:titlesonly:
Report a security issue <report-security-issue>
Harden the cluster <hardening>
```

## Contribute

Contribute to the {{product}} project! Add to the code, documentation or both!

```{toctree}
:titlesonly:
Contribute <contribute>
```

---

## Other documentation types

Our [Reference section] is for when you need to check specific details or
information such as the command reference or release notes.

Alternatively, the [Tutorials section] contains step-by-step tutorials to help
guide you through exploring and using {{product}}.

For a better understanding of how {{product}} works and related topics
such as security, our [Explanation section] helps you expand your knowledge
and get the most out of Kubernetes.

<!--LINKS -->
[Tutorials section]: ../tutorial/index
[Explanation section]: ../explanation/index
[Reference section]: ../reference/index
