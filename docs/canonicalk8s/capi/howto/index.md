# How-to guides

If you have a specific goal, but are already familiar with Kubernetes, our
How-to guides are more specific and contain less background information.
They’ll help you achieve an end result but may require you to understand and
adapt the steps to fit your specific requirements.

```{toctree}
:hidden:
Overview <self>
```

## Install and provision

```{toctree}
:glob:
:titlesonly:
Provision a Canonical Kubernetes cluster <provision>
Install custom Canonical Kubernetes <custom-ck8s>
Use custom bootstrap configuration <custom-bootstrap-config>
```

## Upgrade

Perform important cluster maintenance by upgrading the Kubernetes version and
more.

```{toctree}
:glob:
:titlesonly:
Upgrade the Kubernetes version <rollout-upgrades>
Perform an in-place upgrade <in-place-upgrades>
Upgrade the providers of a management cluster <upgrade-providers>
```

## Certificates

```{toctree}
:glob:
:titlesonly:
Refresh workload cluster certificates <refresh-certs>
Use intermediate CAs with Vault <intermediate-ca.md>
```

## External datastore

```{toctree}
:glob:
:titlesonly:
Use external etcd <external-etcd.md>
```

## Cluster migration

Migrate your management cluster to a different substrate.

```{toctree}
:glob:
:titlesonly:
Migrate the management cluster <migrate-management>
```

## Troubleshoot

Debug issues in your cluster.

```{toctree}
:glob:
:titlesonly:
Troubleshooting <troubleshooting>
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
