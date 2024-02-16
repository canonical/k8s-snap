# Channels

Canonical Kubernetes uses the concept of `channels` to make sure you always get
the version of Kubernetes you are expecting, and that future upgrades can be
handled without minimum, if any, disruption.

## Choosing the right channel

When installing or updating Canonical Kubernetes you can (and should in most
cases) specify a channel. The channel specified is made up of two components;
the **track** and the **risk level**. 

The track will match the minor version of upstream Kubernetes. For example,
specifying the `1.30` track will match upstream releases of the same minor
version ("1.30.0", "1.30.1", "1.30.x" etc.). Releases of Charmed Kubernetes
closely follow the upstream releases and usually follow within 24 hours.

The 'risk level' component of the channel is one of the following:

- **`stable`**: Matches upstream stable releases. 
- **`candidate`**: Tracks upstream release candidate
- **`beta`**: Tracks upstream beta releases - expect bugs
- **`edge`**: Experimental release including upstream alpha releases

Note that for each track, not all risk levels are guranteed to be available.
For example, there may be a new upstream version in devlopment which only has
an `edge` level. For a mature release, there may no longer be any `beta` or
`edge`. In these cases, if you specify a risk level which has no releases for
that track the snap system will choose the closest available release with a
lower risk level. Whatever risk level specified is the **maximum** risk level
of the snap that will be installed - if you choose `candidate` you will never
get `edge` for example.

For all snaps, you can find out what channels are available by running the
`info` command, For example:

```bash
sudo snap info k8s
```

## Updates and switching channels

Updates for upstream patch releases will happen automatically by default. For example, if you have selected the channel `1.30/stable`, your snap will refresh itself on the usual snap [refresh schedule]. These updates should not effect the operation of Canonical Kubernetes.

## FAQ: Which channel is right for me?


I want to always be on the latest release matching a specific upstream K8s release (reccommended).
– Use --channel=<release>/stable, eg --channel=1.25/stable.

I want to test-drive a pre-stable release.

– Use --channel=<next_release>/edge for alpha releases.

– Use --channel=<next_release>/beta for beta releases.

– Use --channel=<next_release>/candidate for candidate releases.

I am waiting for a bug fix on MicroK8s:

– Use --channel=<release>/edge.

I am waiting for a bug fix on upstream Kubernetes:

– Use --channel=<release>/candidate.


Changing channels
It is possible to change the snap channel using the refresh command. E.g. to transition to the latest alpha:

sudo snap refresh microk8s --channel=latest/edge
Changing the refresh schedule
By default, snaps are set to check for updates and automatically refresh to the
latest version (for your selected channel) four times per day. For deployments where this behavior is undesirable you are given the option to postpone, schedule or even block automatic updates. The snap refreshes page outlines how to configure these options.

<!-- LINKS -->

[Snapcraft documentation]: https://snapcraft.io/docs/channels
[refresh schedule]: 