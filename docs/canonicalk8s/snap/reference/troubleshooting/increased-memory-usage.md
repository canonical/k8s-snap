# Increased memory usage in Dqlite

## Problem

The datastore used for {{product}} Dqlite, reported an [issue #196] of increased
memory usage over time. This was particularly evident in smaller clusters.

## Explanation

This issue was caused due to an inefficient resource configuration of
Dqlite for smaller clusters. The threshold and trailing parameters are
related to Dqlite transactions and must be adjusted. The threshold is
the number of transactions we allow before a snapshot is taken of the
leader. The trailing is the number of transactions we allow the follower
node to lag behind the leader before it consumes the updated snapshot of the
leader. Currently, the default snapshot configuration is 1024 for the
threshold and 8192 for trailing which is too large for small clusters. Only
setting the trailing parameter in a configuration yaml automatically sets the
threshold to 0. This leads to a snapshot being taken every transaction and
increased CPU usage.

## Solution

Apply a tuning.yaml custom configuration to the Dqlite datastore in order to
adjust the trailing and threshold snapshot values. The trailing parameter
should be twice the threshold value. Create the tuning.yaml
file and place it in the Dqlite directory
`/var/snap/k8s/common/var/lib/k8s-dqlite/tuning.yaml`:

```
snapshot:
  trailing: 1024
  threshold: 512
```

Restart Dqlite:

```
sudo snap restart snap.k8s.k8s-dqlite
```

<!-- LINKS -->

[issue #196]: https://github.com/canonical/k8s-dqlite/issues/196#issuecomment-2621527026
