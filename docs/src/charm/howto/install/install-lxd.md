# Installing to localhost/LXD

The main [install instructions][install] cover most situations for installing
{{product}} using a charm. However, there are two scenarios which
require some extra steps. These are:

- deploying to the 'localhost' cloud
- deploying to a container on a machine (i.e. when installing a bundle or using
  the 'to:' directive to install to an existing machine)

The container running the charm, or more accurately, the LXD instance
controlling the container, needs to have a particular configuration in order
for the Kubernetes components to operate properly.


## Apply the {{product}} LXD profile

On the machine running the 'localhost' cloud, we can determine the existing
profiles by running the command:

```
lxc profile list
```

For example, suppose we have created a model called `myk8s`. This will
output a table like this:

```
+-----------------+---------------------+---------+
|      NAME       |     DESCRIPTION     | USED BY |
+-----------------+---------------------+---------+
| default         | Default LXD profile | 2       |
+-----------------+---------------------+---------+
| juju-controller |                     | 1       |
+-----------------+---------------------+---------+
| juju-myk8s      |                     | 0       |
+-----------------+---------------------+---------+
```

Each model created by Juju will generate a new profile for LXD. We can inspect
and edit the profiles easily by using `lxc` commands.

## Fetching the profile

A working LXD profile is kept in the source repository for the {{product}}
'k8s' snap. You can retrieve this profile by running the command:

<!-- markdownlint-disable -->
```
wget https://raw.githubusercontent.com/canonical/k8s-snap/main/tests/integration/lxd-profile.yaml -O k8s.profile
```
<!-- markdownlint-restore -->

To pipe the content of the file into the k8s LXD profile, run:

```
cat k8s.profile | lxc profile edit juju-myk8s
```

Remove the copied content from your directory:

```
rm k8s.profile
```

The profile editor will syntax-check the profile as part of the editing
process, but you can confirm the contents have changed by running:

```
lxc profile show juju-myk8s
```

```{note} For an explanation of the settings in this file,
   [see below](explain-rules-charm)
```

## Deploying to a container

We can now deploy {{product}} into the LXD-based model as described in
the [charm][] guide.

(explain-rules-charm)=

## Explanation of custom LXD rules

**boot.autostart: “true”**: Always start the container when LXD starts. This is
needed to start the container when the host boots.

**linux.kernel_modules**: Comma separated list of kernel modules to load before
starting the container

**lxc.apparmor.profile=unconfined**: Disable [AppArmor]. Allow the container to
talk to a bunch of subsystems of the host (e.g. `/sys`). By default AppArmor
will block nested hosting of containers, however Kubernetes needs to host
Containers. Containers need to be confined based on their profiles thus we rely
on confining them and not the hosts. If you can account for the needs of the
Containers you could tighten the AppArmor profile instead of disabling it
completely, as suggested in S.Graber's notes[^1].

**lxc.cap.drop=**: Do not drop any capabilities [^2]. For justification see
above.

**lxc.mount.auto=proc:rw sys:rw**: Mount proc and sys rw. For privileged
containers, lxc over-mounts part of /proc as read-only to avoid damage to the
host. Kubernetes will complain with messages like `Failed to start
ContainerManager open /proc/sys/kernel/panic: permission denied`

**lxc.cgroup.devices.allow=a**: "a" stands for "all." This means that the
container is allowed to access all devices. It's a wildcard character
indicating permission for all devices. For justification see above.

**security.nesting: “true”**: Support running LXD (nested) inside the
container.

**security.privileged: “true”**: Runs the container in privileged mode, not
using kernel namespaces [^3], [^4]. This is needed because hosted Containers may
need to access for example storage devices (See comment in [^5]).

<!-- LINKS -->
<!-- markdownlint-disable MD034 -->
[^1]: https://stgraber.org/2012/05/04/
[^2]: https://stgraber.org/2014/01/01/lxc-1-0-security-features/
[^3]: https://unix.stackexchange.com/questions/177030/what-is-an-unprivileged-lxc-container/177031#177031
[^4]: http://blog.benoitblanchon.fr/lxc-unprivileged-container/
[^5]: https://wiki.ubuntu.com/LxcSecurity

[AppArmor]: https://apparmor.net/
[charm]:    ./charm
[install]:  ./charm
