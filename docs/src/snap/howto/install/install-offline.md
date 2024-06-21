# Installing Canonical Kubernetes Offline or in an airgapped environment

There are situations where it is necessary or desirable to run Canonical Kubernetes on a
machine that is not connected to the internet. 
Based on different degrees of separation from the network, different solutions are offered to accomplish this goal.
This guide explains the necessary preparation required for the offline installation and walks you through the different potential scenarios.

# Install Canonical Kubernetes in air gapped environments

## Prepare for Deployment

In preparation for the offline deployment you will download the Canonical Kubernetes snap, fulfill the networking requirements based on your scenario and handle images for workloads and Canonical Kubernetes features.

### 1. Download the Canonical Kubernetes snap

From a machine with access to the internet download the following:
```
sudo snap download k8s --channel 1.30-classic/beta
sudo snap download core20
sudo mv k8s_*.snap k8s.snap
sudo mv k8s_*.assert k8s.assert
sudo mv core20_*.snap core20.snap
sudo mv core20_*.assert core20.assert
```

The [core20][Core20] and `k8s` snap are downloaded. The `core20.assert` and `k8s.assert` files, are necessary to verify the integrity of the snap packages.

```{note} Update the version of k8s by adjusting the channel parameter. Find the version you desire in the [snapstore][snapstore].
```




<!-- LINKS -->

[Getting started]: getting-started
[Core20]: https://canonical.com/blog/ubuntu-core-20-secures-linux-for-iot
[snapstore]: https://snapcraft.io/k8s