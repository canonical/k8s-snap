# Cilium pod fails to start as `cilum_vxlan: address already in use`

## Problem

When deploying {{product}} the Cilium pods fail to start and reports the error:

```
failed to start: daemon creation failed: error while initializing daemon: failed
while reinitializing datapath: failed to setup vxlan tunnel device: setting up
vxlan device: creating vxlan device: setting up device cilium_vxlan: address
already in use
```

## Explanation

Fan networking is automatically enabled in some substrates. This causes
conflicts with some CNIs such as Cilium. This conflict of
`address already in use` causes Cilium to be unable to set up it's VXLAN
tunneling network. There may also be other networking components on the system
attempting to use the default port for their own VXLAN interface that will
cause the same error.

## Solution

Configure Cilium to use another tunnel port. Set the annotation `tunnel-port`
to an appropriate value (the default is 8472).

```
sudo k8s set annotation="k8sd/v1alpha1/cilium/tunnel-port=<PORT-NUMBER>"
```

Since the Cilium pods are in a failing state, the recreation of the VXLAN
interface is automatically triggered. Verify the VXLAN interface has come
up:

```
ip link list type vxlan
```

It should be named `cilium_vxlan` or something similar.

Verify that Cilium is now in a running state:

```
sudo k8s kubectl get pods -n kube-system
```