# Installing to localhost/LXD

The main [install instructions][install] work for most circumstances when you
want to install Canonical Kubernetes from a charm. There are a couple of
scenarios which require some extra steps however. These are:

- deploying to the 'localhost' cloud
- deploying to a container on a machine (i.e. when installing a bundle or using
  the 'to:' directive to install to an existing machine)

The container running the charm, or more accurately, the LXD instance
controlling the container, needs to have a particular configuration in order
for the Kubernetes components to operate properly.

## Fetching the profile

A working LXD profile is kept in the source repository for the Canonical
Kubernetes 'k8s' snap. You can retreive this profile by running the command:

<!-- markdownlint-disable -->
```
wget https://raw.githubusercontent.com/canonical/k8s-snap/main/tests/integration/lxd-profile.yaml -O k8s.profile
```
<!-- markdownlint-restore -->

## Applying the profile to the localhost cloud

On the machine running the 'localhost' cloud, we can determine the existing
profiles by running the command:

```
lxc profile list
```

For example, suppose we have created a model called 'myk8s'. This will
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

To replace the model's profile with the Kubernetes-specific one we downloaded,
run the command:

```
cat k8s.profile | lxc profile edit juju-myk8s
```

The profile editor will syntax-check the profile as part of the editing
process, but you can confirm the contents have changed by running:

```
lxc profile show juju-myk8s
```

```{note} You need to change this profile ***before*** deploying any charms!
```

## Deploying to a container
