# Upgrade notes

```{warning}
You are currently on version `release-1.32` of our docs. Reminder to switch 
to the version of the docs you aim to upgrade to in order to see the release 
notes and get the latest updates for that version. 
```

## Upgrade 1.32 to 1.33

If you are not using dual stack networking, you can simply run:

```bash
sudo snap refresh k8s --channel=1.33/stable
```

All components will be updated automatically.

### Additional steps for dual-stack environments

If your cluster is configured with dual stack networking (IPv4 and IPv6), 
you’ll need to make a manual adjustment before refreshing. {{product}} 1.33 
includes Cilium v1.17, which introduces a stricter requirement for dual stack: 
each node must report both IPv4 and IPv6 addresses to the API server. 
If this is not satisfied, the Cilium agent pods will fail to start. 
For each node in the cluster:

- Update the `--node-ip` flag in the kubelet configuration file 
`/var/snap/k8s/common/args/kubelet` to include both the IPv4 and IPv6 addresses 
(comma-separated) from the network interface that is used to connect the node 
to the cluster network:

```bash
--node-ip=<IPv4>,<IPv6>
```

- Restart the `kubelet` service

```bash
sudo systemctl restart snap.k8s.kubelet.service
```

- Restart the Cilium DaemonSet:

```bash
sudo k8s kubectl rollout restart daemonset cilium -n kube-system
```

Now you can run the snap `refresh` command to perform the upgrade.

### Verify the upgrade 

Check the `k8s` snap version has been updated and the cluster is back in the 
`Ready` state.

```
snap info k8s
sudo k8s status --wait-ready
```


