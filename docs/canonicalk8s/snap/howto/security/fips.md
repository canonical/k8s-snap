# How to set up a FIPS-compliant Kubernetes cluster

FIPS (Federal Information Processing Standards) ensures security compliance
crucial for US government and regulated industries. This how-to guide provides
the steps to set up a FIPS-compliant Kubernetes cluster using the {{ product }} snap.

## Enable FIPS on an Ubuntu host machine

To enable FIPS on your host machine, you require an [Ubuntu Pro] subscription.
Open the [Ubuntu Pro subscription dashboard] to retrieve your Ubuntu Pro token
required to enable access to FIPS-certified modules on your system.

``` {note}
If this section leaves open any further questions consult the [enable FIPS with Ubuntu]
guide for more detailed instructions.
```

As a first step, ensure that your Ubuntu Pro Client is installed and running at
least 27.0:

```
pro version
27.13.5~16.04.1
```

If you have not installed the Ubuntu Pro Client yet or have an older version,
please run:

```
sudo apt update
sudo apt install ubuntu-advantage-tools
```

Attach the Ubuntu Pro token with the `--no-auto-enable` option to prevent
Canonical Livepatch services, which are not supported with FIPS:

```
sudo pro attach <your_pro_token> --no-auto-enable
```

Now, enable the FIPS crypto modules on your host machines:

```
sudo pro enable fips-updates
```

Reboot your host machine to apply the changes:

```
sudo reboot
```

Verify your host machine is running in FIPS mode:

```
cat /proc/sys/crypto/fips_enabled
```

If the output is `1`, your host machine is running in FIPS mode.

## Firewall configuration for Kubernetes

{{ product }} requires certain firewall rules and guidelines to ensure its operation.
Additionally, please review your services hosted in Kubernetes and add any
necessary firewall rules.

The following rules are recommended for a {{ product }} cluster with FIPS enabled.

```{warning}
The presented network rules may be incompatible with your network setup, or you
may find them too permissive or too restrictive. Please review and adjust them
according to your network requirements.
```

### Enable package forwarding

Forwarding is needed as containers typically live in isolated networks and need
the host to route traffic between their internal network and the outside world.

Configure your firewall (UFW) to allow packet forwarding by editing `/etc/default/ufw`:

```
DEFAULT_FORWARD_POLICY="ACCEPT"
```

Then, enable IP forwarding in the kernel by editing `/etc/sysctl.conf`:

```
net.ipv4.ip_forward=1
```

Alternatively, apply the change immediately with:

```
sudo sysctl -w net.ipv4.ip_forward=1
```

As a last step, reload the firewall rules:

```
sudo ufw reload
```

### Allow access to the Kubernetes services

For detailed information about the ports and services used by {{ product }},
see the [ports-and-services] documentation.

Allow the following ports in your firewall:

```
sudo ufw allow 10250/tcp # kubelet
sudo ufw allow 10257/tcp # kube-controller
sudo ufw allow 10259/tcp # kube-scheduler
sudo ufw allow 6443/tcp # coredns
sudo ufw allow 2379/tcp # etcd
sudo ufw allow 2380/tcp # etcd
sudo ufw allow 6400/tcp # k8sd
sudo ufw allow 4240/tcp # cilium-agent
sudo ufw allow 8472/udp # cilium-agent
```

### Ensure runtime with FIPS-certified libraries

Install the [core22] runtime with FIPS-certified libraries. The core22 snap
offers the fips-updates track, which contains NIST-certified packages along
with [security patches].

```
sudo snap install core22 --channel=fips-updates/stable
```

In case you have core22 already installed, perform a snap refresh to update it
to the latest version:

```
sudo snap refresh core22 --channel=fips-updates/stable
```

### Install Canonical Kubernetes

Inststall {{ product }} on your fips host:

```
sudo snap install k8s --channel=1.32-classic/candidate/fips-early-release --classic
```
<!-- TODO: Update once FIPS is in stable -->
```{warning}
This command installs the Kubernetes snap from the FIPS-enabled candidate
channel. Please note that this is an early release; only Kubernetes services are
FIPS-compliant, not the additional features (which are OCI images deployed
separately when bootstrapping). Once FIPS is fully supported, the channel will
be updated to stable.
```

The snap includes binaries built with FIPS-compliant cryptography. The
components will automatically detect if the system is running in FIPS mode and
activate internal FIPS-related settings accordingly.

After the snap installation completes, you can bootstrap the node as usual:

```
sudo k8s bootstrap
```

Then you may wait for the node to be ready, by running:

```
sudo k8s status
```

Your FIPS-compliant Kubernetes cluster is now ready for workload deployment and
additional node integrations. Please ensure that your workloads are
FIPS-compliant as well, to maintain the security standards required by FIPS.

<!-- LINKS -->
[Ubuntu Pro]: https://ubuntu.com/pro
[Ubuntu Pro subscription dashboard]: https://ubuntu.com/pro/dashboard
<!-- markdownlint-disable MD053 -->
[enable FIPS with Ubuntu]: https://ubuntu.com/tutorials/using-the-ubuntu-pro-client-to-enable-fips#1-overview
<!-- markdownlint-enable MD053 -->
[core22]: https://snapcraft.io/core22
[security patches]: <https://ubuntu.com/security/certifications/docs/16-18/fips-updates>
[ports-and-services]: ../reference/ports-and-services
