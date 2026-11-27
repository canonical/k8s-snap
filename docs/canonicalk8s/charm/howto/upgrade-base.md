---
myst:
  html_meta:
    description: "How to upgrade the machines running the k8s and k8s-worker charms from Ubuntu 22.04 LTS to Ubuntu 24.04 LTS in place, before upgrading to 1.36."
---

# How to upgrade the Ubuntu base of {{product}} charm units

From 1.36, the `k8s` and `k8s-worker` charms no longer support Ubuntu 22.04
LTS. If your cluster runs on Ubuntu 22.04 LTS, move its machines to Ubuntu 24.04
LTS before you upgrade to 1.36.

This guide upgrades each machine in place with Juju while the cluster stays on
1.35. Each node is drained before its machine is upgraded, so its workloads
move to the other nodes.

This guide covers the machines that run the `k8s` and `k8s-worker` charms.
Other applications in the model, such as an external `etcd` cluster, need
their own base upgrade.

## Prerequisites

- A {{product}} cluster deployed with the `k8s` and `k8s-worker` charms on the
  `1.35/stable` channel. If you are on an earlier version, first follow the
  [minor upgrade steps][minor-upgrades] one version at a time until you reach
  1.35.
- Juju 3.6 for the controller, the model and the client. Juju 4.0 removes the
  `upgrade-machine` command used in this guide.
- `kubectl` access to the cluster. To get a kubeconfig, run
  `juju run k8s/leader get-kubeconfig`.
- At least three `k8s` units, so the Kubernetes API and the datastore stay
  available while one control-plane machine is down. With a single `k8s`
  unit, the API is unavailable while that machine is upgraded.
- Enough spare capacity on the other nodes to run the workloads of the node
  you drain.
- A recent backup of your cluster data and workloads. See the
  [backup and restore guide][backup-restore].

```{warning}
`juju upgrade-machine` cannot be undone. Test this procedure on a
non-production cluster first.
```

## Understand what happens during the upgrade

Each charm publishes a separate build for each Ubuntu release. Units
deployed on Ubuntu 22.04 LTS run the 22.04 build, which can't run on
Ubuntu 24.04 LTS. This has three effects:

- Juju refuses to prepare the machine unless you add `--force`.
- After a machine is upgraded, the charm on that unit stops working and the
  unit shows an `error` status. The {{product}} services on the machine,
  including the Kubernetes API server, kubelet and the datastore, keep
  running.
- The charm works again after you refresh the application to its Ubuntu
  24.04 LTS build. You can only do this after **every** machine of that
  application runs Ubuntu 24.04 LTS.

Until an application is refreshed, Juju can't manage its upgraded units.
During that time, don't change charm configuration, integrations or the
number of units.

## Plan the upgrade order

Upgrade one machine at a time, and finish one application before you start
the next:

1. Upgrade every `k8s-worker` machine, then
   [refresh](#refresh-the-application-to-its-ubuntu-2404-lts-build) `k8s-worker`.
   If your model has several `k8s-worker` applications, do this for each
   one.
2. Upgrade every `k8s` machine, leaving the leader until last, then refresh
   `k8s`. `juju status` marks the leader unit with `*`.

Each machine can take 30 minutes or longer. Most of that time is spent in
`do-release-upgrade`.

## Upgrade a machine

Repeat these steps for each machine. Set `M` to the Juju machine number from
`juju status`, and `UNIT` to the unit on that machine:

```bash
M=3
UNIT=k8s-worker/0
```

### Check the cluster

Before each machine, confirm that every node is `Ready` and every pod is
`Running` or `Completed`:

```bash
kubectl get nodes
kubectl get pods -A
```

If your kubeconfig points at the control-plane machine you are about to
upgrade, point it at another control-plane machine for now. Find the cluster
name, then change its server address:

```bash
kubectl config get-clusters
kubectl config set-cluster <cluster-name> --server=https://<other-control-plane-ip>:6443
```

### Drain the node

The node name is the machine's hostname in lowercase:

```bash
NODE=$(juju ssh "$M" -- hostname | tr -d '\r' | tr '[:upper:]' '[:lower:]')
kubectl drain "$NODE" --ignore-daemonsets --delete-emptydir-data
```

The charms don't drain nodes during a base upgrade, so always drain first.
If a node isn't drained, Services can send traffic to its pods while the
machine reboots.

### Update Ubuntu 22.04 LTS

`do-release-upgrade` only starts when all available updates are installed:

```bash
juju ssh "$M" -- 'sudo apt-get update && sudo DEBIAN_FRONTEND=noninteractive apt-get -y \
  -o Dpkg::Options::=--force-confdef -o Dpkg::Options::=--force-confold full-upgrade'
juju ssh "$M" -- sudo systemctl reboot
```

The reboot command may end with a connection error. This is expected, because
the machine closes the SSH session as it shuts down. Wait until the machine
accepts SSH again before you continue:

```bash
sleep 20
until juju ssh "$M" -- true 2>/dev/null; do sleep 10; done
```

### Prepare the machine

```bash
juju upgrade-machine "$M" prepare ubuntu@24.04 --force
```

Without `--force`, Juju stops with an error like this:

```text
ERROR charm "k8s-worker" does not support ubuntu@24.04/stable, force not used
```

### Upgrade to Ubuntu 24.04 LTS

Run `do-release-upgrade` non-interactively and keep the existing configuration
files. Running it as a systemd unit lets it continue if your SSH connection
drops:

```bash
juju ssh "$M" -- 'printf "%s\n" "DPkg::Options { \"--force-confdef\"; \"--force-confold\"; };" |
    sudo tee /etc/apt/apt.conf.d/99-release-upgrade >/dev/null &&
  sudo systemd-run --unit=release-upgrade -r --setenv=DEBIAN_FRONTEND=noninteractive \
    do-release-upgrade -f DistUpgradeViewNonInteractive'
```

Check progress until the state is `active/exited`. If it shows `failed`, see
[Troubleshooting](#troubleshooting).

```bash
juju ssh "$M" -- 'systemctl show -p ActiveState -p SubState --value release-upgrade | paste -sd/'
```

SSH can be unavailable for several minutes during the upgrade. The node keeps
serving its workloads during that time.

When the upgrade has finished, reboot the machine and wait for it to accept SSH
again:

```bash
juju ssh "$M" -- sudo systemctl reboot
sleep 20
until juju ssh "$M" -- true 2>/dev/null; do sleep 10; done
```

Then confirm the release:

```bash
juju ssh "$M" -- lsb_release -cs
```

The output should be `noble`.

Remove the temporary apt setting, so later package upgrades handle
configuration files normally again:

```bash
juju ssh "$M" -- sudo rm /etc/apt/apt.conf.d/99-release-upgrade
```

### Complete the upgrade in Juju

```bash
juju upgrade-machine "$M" complete
```

`complete` waits for the unit's `post-series-upgrade` hook. That hook fails
because the 22.04 build of the charm can't run on Ubuntu 24.04 LTS. The unit
shows `hook failed: "post-series-upgrade"`, and `juju debug-log` shows
`ModuleNotFoundError`.

In a second terminal, skip the failed hook:

```bash
juju resolve --no-retry "$UNIT"
```

`complete` then finishes. The charm has no work to do in this hook, so
skipping it is safe.

```{note}
Use `juju resolve` only for the `post-series-upgrade` hook. Later hooks on
this unit may also fail until you refresh the application. Leave them; the
refresh resolves them.
```

### Return the node to service

```bash
kubectl uncordon "$NODE"
kubectl get nodes
kubectl get pods -A
```

The node should be `Ready` and the pods healthy. `juju status` still shows the
unit in `error` until you refresh the application.

For a `k8s` machine, also check the cluster from another control-plane
machine:

```bash
juju ssh <other-control-plane-machine> -- sudo k8s status
```

Every control-plane node should be listed and the cluster should be `ready`.

## Refresh the application to its Ubuntu 24.04 LTS build

When every machine of an application shows `ubuntu@24.04` in the `Base`
column of `juju status`, refresh the application on its current channel:

```bash
APP=k8s-worker
juju refresh "$APP" --channel 1.35/stable --base ubuntu@24.04 --force-units --force-base
juju status "$APP" --watch 10s
```

| Option | Purpose |
|---|---|
| `--base ubuntu@24.04` | Selects the Ubuntu 24.04 LTS build of the charm |
| `--force-units` | Refreshes units that are in `error` |
| `--force-base` | Allows the refresh while Juju still records the application's base as `ubuntu@22.04` |

`juju set-application-base` can't be used instead, because it rejects a base
that the deployed build doesn't support. The refresh updates the application
base for you.

After the refresh, every unit should return to `active`/`idle` without any
further action. The `k8s` leader may show `Waiting to bootstrap k8s snap` for
a few minutes after the refresh. This clears by itself, so don't run any
action to fix it.

Then repeat the machine upgrades and the refresh for the next application.

## Verify the base upgrade

```bash
juju status
kubectl get nodes -o wide
kubectl get pods -A
```

Check that:

- Every machine shows `ubuntu@24.04` in the `Base` column.
- Every unit is `active`/`idle`.
- Every node is `Ready` and shows `Ubuntu 24.04` in the `OS-IMAGE` column.
- Every pod is `Running` or `Completed`.

To check the cluster further, run a [cluster validation][cluster-validation].

You can now [upgrade to 1.36][upgrade-136].

## Troubleshooting

**`do-release-upgrade` failed.** Read the upgrade logs on the machine:

```bash
juju ssh "$M" -- 'sudo journalctl -u release-upgrade --no-pager | tail -40'
juju ssh "$M" -- 'sudo tail -30 /var/log/dist-upgrade/main.log'
juju ssh "$M" -- 'sudo tail -30 /var/log/dist-upgrade/term.log'
```

**`do-release-upgrade` doesn't find Ubuntu 24.04 LTS.** Check that
`/etc/update-manager/release-upgrades` on the machine sets `Prompt=lts`:

```bash
juju ssh "$M" -- grep '^Prompt' /etc/update-manager/release-upgrades
```

**`juju upgrade-machine complete` doesn't return.** It is waiting for the
failed `post-series-upgrade` hook. Run `juju resolve --no-retry "$UNIT"`.

**`juju refresh` fails with `base "ubuntu@22.04" not supported by charm`.**
Check that every machine of the application runs Ubuntu 24.04 LTS, then run the
refresh with all three options shown above.

**A unit stays in `error` after the refresh.** Check in `juju status` that
the `Rev` column changed, then look for errors with
`juju debug-log --include "$UNIT"`.

<!-- LINKS -->

[minor-upgrades]: ./upgrade.md#minor-upgrades
[upgrade-136]: ./upgrade.md#upgrade-from-135-to-136
[backup-restore]: ../../snap/howto/backup-restore
[cluster-validation]: ./validate
