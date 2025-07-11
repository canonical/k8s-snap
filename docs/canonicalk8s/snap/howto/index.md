# How-to guides

If you have a specific goal, but are already familiar with Kubernetes, our
How-to guides are more specific and contain less background information.
They’ll help you achieve an end result but may require you to understand and
adapt the steps to fit your specific requirements.

```{toctree}
:hidden:
Overview <self>
```

## Installation

Installation follows a very similar pattern on all platforms, but some minor
differences must be addressed in each case. You may also want to customize the
installation of your {{product}} nodes.

```{toctree}
:titlesonly:
Install <install/index>
```

## Networking

{{product}} comes with default networking features for a fully functioning
cluster. More advanced features can be enabled by a few configuration steps.

```{toctree}
:titlesonly:
networking/index
```

## Storage

Specific storage needs of your cluster can be met by setting up persistent
storage or replacing the default datastore with an external one such as `etcd`.

```{toctree}
:titlesonly:
storage/index
Use an external datastore <external-datastore>
```

## Security and compliance

Harden your cluster according to industry standards.

```{toctree}
:titlesonly:
security/index
```

## Cluster upgrades and refreshes

```{toctree}
:titlesonly:
Manage upgrades <upgrades>
Refresh Kubernetes Certificates <refresh-certs>
Use intermediate CAs with Vault <intermediate-ca.md>
```

## Manage images

Manage cluster images directly through containerd.

```{toctree}
:titlesonly:
Manage images <image-management.md>
```

## Cluster back up and restore

```{toctree}
:titlesonly:
Back up and restore <backup-restore>
```

## Monitoring and troubleshooting

Sometimes things go wrong and you need to troubleshoot. Having observability
set up on your cluster can greatly increase the rate at which a problem is
identified and solved.

```{toctree}
:titlesonly:
Set up cluster observability  <observability>
Recover a cluster after quorum loss <restore-quorum>
Troubleshooting <troubleshooting>
Get support <support>
```

## Enhanced Platform Awareness

EPA utilizes server hardware capabilities in the {{product}} cluster. It
exposes technologies such as HugePages, CPU pinning, SR-IOV and more.

```{toctree}
:titlesonly:
Set up Enhanced Platform Awareness <epa>
```

## Contribute

Contribute to the {{product}} project! Add to the code, documentation or both!

```{toctree}
:titlesonly:
Contribute <contribute>
```

---

## Other documentation types

Our Reference section is for when you need to check specific details or
information such as the command reference or release notes.

Alternatively, the [Tutorials section] contains step-by-step tutorials to help
guide you through exploring and using {{product}}.

For a better understanding of how {{product}} works and related topics
such as security, our [Explanation section] helps you expand your knowledge
and get the most out of Kubernetes.

Finally, our [Reference section] is for when you need to check specific details
or information such as the command reference or release notes.

<!--LINKS -->
[Tutorials section]: ../tutorial/index
[Explanation section]: ../explanation/index
[Reference section]: ../reference/index
