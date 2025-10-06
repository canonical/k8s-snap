# How to set up a FIPS compliant Kubernetes cluster

[FIPS 140-3] (Federal Information Processing Standards) ensures security
compliance crucial for US government and regulated industries. This
how-to guide provides the steps to set up a FIPS compliant Kubernetes
cluster using the {{ product }} snap.

## Enable FIPS on an Ubuntu host machine

To enable FIPS on your host machine, you require an [Ubuntu Pro] subscription.
Open the [Ubuntu Pro subscription dashboard] to retrieve your Ubuntu Pro token
required to enable access to FIPS-certified modules on your system.

Ensure that your Ubuntu Pro Client is installed and running at
least 27.0:

```
pro version
```

If you have not installed the Ubuntu Pro Client yet or have an older version,
run:

```
sudo apt update
sudo apt install ubuntu-advantage-tools
```

Attach the Ubuntu Pro token with the `--no-auto-enable` option to prevent
Canonical Livepatch services, which are not supported with FIPS:

```
sudo pro attach <your_pro_token> --no-auto-enable
```

Now, enable the FIPS crypto modules on your host machine:

```
sudo pro enable fips-updates
```

Reboot to apply the changes:

```
sudo reboot
```

Verify your host machine is running in FIPS mode:

```
cat /proc/sys/crypto/fips_enabled
```

If the output is `1`, your host machine is running in FIPS mode.

``` {note}
If this section leaves open any further questions consult the [enable FIPS with Ubuntu]
guide for more detailed instructions.
```

## Firewall configuration for Kubernetes

{{ product }} requires certain firewall rules and guidelines to
ensure its operation. Additional firewall rules may also be necessary based on
user deployed workloads and services. Please follow the steps in the
[firewall configuration] guide.

## Ensure runtime with FIPS-certified libraries

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

## Install Canonical Kubernetes

Install {{ product }} on your FIPS host:

```
sudo snap install k8s --classic
```

```{note}
Please note that FIPS is only available in the `k8s` release 1.34 and later.
If you are using an earlier version, you will need to upgrade to the latest
version of the snap to use FIPS support.
```

The k8s snap can leverage the host's FIPS compliant
cryptography. The components will automatically detect if the system is
running in FIPS mode and activate internal FIPS-related settings
accordingly.

TODO reword
```{attention}
If you intend to apply DISA STIG hardening to your cluster, go to the DISA STIG deployment guide to get detailed steps on deploying with a strciter config bootstrap file
```

After the snap installation completes, you can bootstrap the node as usual:

```
sudo k8s bootstrap
```

Then you may wait for the node to be ready, by running:

```
sudo k8s status
```

Your Kubernetes cluster is now ready for workload deployment and
additional node integrations. Please ensure that your workloads and
underlying system and hardware are FIPS compliant as well, to
maintain the security standards required by FIPS. For example,
ensure that your container images used for your applications can
be used with the hosts FIPS compliant libraries.


## Disable FIPS on an Ubuntu host machine

```{warning}
Disabling FIPS on a host machine is not recommended: only
enable FIPS on machines intended expressly to be used for FIPS.
Changing the FIPS mode may have implications for the
services running on your live cluster, so ensure you understand the
consequences of disabling FIPS before proceeding.
```

To disable FIPS on your host machine, run the following command:

```
sudo pro disable fips-updates
```

For further information on how to disable FIPS on the host,
consult the [disabling FIPS with Ubuntu] guide.

You can also change the [core22] snap back to the default
non-FIPS channel:

```
sudo snap refresh core22 --channel=latest/stable
```

Then reboot your host machine to apply the changes:

```
sudo reboot
```

After the reboot, the k8s snap's k8sd service will restart and
automatically detect that the host is no longer in FIPS mode
and will revert to the default non-FIPS settings.

<!-- LINKS -->
[FIPS 140-3]: https://csrc.nist.gov/pubs/fips/140-3/final
[Ubuntu Pro]: https://ubuntu.com/pro
[Ubuntu Pro subscription dashboard]: https://ubuntu.com/pro/dashboard
<!-- markdownlint-disable MD053 -->
[enable FIPS with Ubuntu]: https://ubuntu.com/tutorials/using-the-ubuntu-pro-client-to-enable-fips#1-overview
<!-- markdownlint-enable MD053 -->
[firewall configuration]: ../networking/ufw
[core22]: https://snapcraft.io/core22
[security patches]: <https://ubuntu.com/security/certifications/docs/16-18/fips-updates>
[disabling FIPS with Ubuntu]: https://documentation.ubuntu.com/pro-client/en/latest/howtoguides/enable_fips/#how-to-disable-fips

