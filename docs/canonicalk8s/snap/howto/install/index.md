---
myst:
  html_meta:
    description: "Canonical Kubernetes snap installation how-to guide index, covering snap, Multipass, LXD, air-gapped, DISA STIG, FIPS, and custom configuration."
---

# Install {{product}}

There's more than one way to install {{product}}. You'll find links to
the current How-to guides below.

```{toctree}
:glob:
:titlesonly:

k8s snap <snap.md>
Customize bootstrap configuration <custom-bootstrap-config>
Add a worker <custom-worker.md>
In development environments <dev-env.md>
With Multipass <multipass>
With LXD <lxd.md>
In air-gapped environments <offline.md>
In FIPS mode <fips.md>
DISA STIG hardened cluster <disa-stig.md>
Uninstall the snap <uninstall.md>
```
