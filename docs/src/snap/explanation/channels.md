# Channels

{{product}} uses the concept of `channels` to make sure you always get
the version of Kubernetes you are expecting, and that future upgrades can be
handled with minimum, if any, disruption.

## Choosing the right channel

When installing or updating {{product}} you can (and should in most
cases) specify a channel. The channel specified is made up of two components;
the **track** and the **risk level**.

The track matches the minor version of upstream Kubernetes. For example,
specifying the `1.30` track will match upstream releases of the same minor
version ("1.30.0", "1.30.1", "1.30.x" etc.). Releases of {{product}}
closely follow the upstream releases and usually follow within 24 hours.

The 'risk level' component of the channel is one of the following:

- **`stable`**: Matches upstream stable releases
- **`candidate`**: Holds the release candidates of the snap
- **`beta`**: Tracks the beta releases - expect bugs
- **`edge`**: Experimental release including upstream alpha releases

Note that for each track, not all risk levels are guaranteed to be available.
For example, there may be a new upstream version in development which only has
an `edge` level. For a mature release, there may no longer be any `beta` or
`candidate`. In these cases, if you specify a risk level which has no releases for
that track the snap system will choose the closest available release with a
lower risk level. Whatever risk level specified is the **maximum** risk level
of the snap that will be installed - if you choose `candidate` you will never
get `edge` for example.

For all snaps, you can find out what channels are available by running the
`info` command, For example:

```
snap info k8s
```

More information can be found in the [Snapcraft documentation][]

## Updates and switching channels

Updates for upstream patch releases will happen automatically by default. For
example, if you have selected the channel `1.30/stable`, your snap will refresh
itself regularly keeping your cluster up-to-date with the latest patches.
For deployments where this behaviour is undesirable you are given the option to
postpone, schedule or even block automatic updates.
The [Snap refreshes documentation] page outlines how to configure these options.

To change the channel of an already installed snap, the `refresh` command can
be used:

```
sudo snap refresh k8s --channel=<new-channel>
```

```{warning}
Changing the channel of an installed snap could result in loss of service. Please
check any release notes or upgrade guides first!
```

## Which channel is right for me?

Choosing the most appropriate channel for your needs depends on a number of
factors. We can give some general guidance for the following cases:

- **I want to always be on the latest stable version matching a specific
upstream K8s release (recommended).**

Specify the release, for example: `--channel=1.30/stable`.

- **I want to test-drive a pre-stable release**

Use `--channel=<next_release>/edge` for alpha releases.

Use `--channel=<next_release>/beta` for beta releases.

Use `--channel=<next_release>/candidate` for candidate releases.

- **I am waiting to test a bug fix on {{product}}**

Use `--channel=<release>/edge`.

- **I am waiting for a bug fix from upstream Kubernetes**

Use `--channel=<release>/candidate`.

<!-- LINKS -->

[Snapcraft documentation]: https://snapcraft.io/docs/channels
[Snap refreshes documentation]: https://microk8s.io/docs/snap-refreshes
