# How to install a FIPS compliant Kubernetes cluster

The [Federal Information Processing Standard] (FIPS) 140-3 is a US government
security standard regulating the use of cryptography. Compliance is crucial for
US government and regulated industries. This how-to guide provides the steps to
set up a FIPS compliant Kubernetes cluster using the
{{ product }} snap.

Please note that FIPS is only available in the `k8s` snap release 1.34 and
later. If you are using an earlier version, you will need to upgrade to
a newer version of the snap to use FIPS mode.

## Prerequisites

This guide assumes the following:

- Ubuntu 22.04 machine with at least 4GB of RAM and 30 GB disk storage
- You have root or sudo access to the machine
- Internet access on the machine

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

## Ensure runtime with FIPS-certified libraries

Install the [core22] runtime with FIPS-certified libraries from the
`fips-updates` track, which contains NIST-certified packages along with
[security patches].

```
sudo snap install core22 --channel=fips-updates/stable
```

If core22 is already installed, a message will be displayed:
`snap "core22" is already installed, see 'snap help refresh'`.
In this case, use the refresh command instead of install.

```
sudo snap refresh core22 --channel=fips-updates/stable
```

## Install {{product}}

Install {{ product }} on your FIPS host:

```{literalinclude} /_parts/install.md
:start-after: <!-- snap start -->
:end-before: <!-- snap end -->
```

The components will automatically detect if the system is
running in FIPS mode and activate internal FIPS-related settings
accordingly.

## Bootstrap the cluster  

```{attention}
If you are deploying a DISA STIG hardened cluster, stop here and instead
continue following the
[Canonical Kubernetes DISA STIG deployment guide](disa-stig.md) to get detailed
instructions on deploying with a stricter bootstrap configuration file or
joining the cluster with a stricter join configuration file.
```

After the snap installation completes, you can bootstrap the node as usual:

```
sudo k8s bootstrap
```

Then you may wait for the node to be ready, by running:

```
sudo k8s status --wait-ready
```

Your Kubernetes cluster is now ready for workload deployment and
You now have a single node Kubernetes cluster operating in FIPS mode and can
add additional nodes or begin deploying workloads.

<!-- LINKS -->
[Federal Information Processing Standard]: https://csrc.nist.gov/pubs/fips/140-3/final
[Ubuntu Pro]: https://ubuntu.com/pro
[Ubuntu Pro subscription dashboard]: https://ubuntu.com/pro/dashboard
<!-- markdownlint-disable MD053 -->
[enable FIPS with Ubuntu]: https://ubuntu.com/tutorials/using-the-ubuntu-pro-client-to-enable-fips#1-overview
<!-- markdownlint-enable MD053 -->
[core22]: https://snapcraft.io/core22
[security patches]: <https://ubuntu.com/security/certifications/docs/16-18/fips-updates>

