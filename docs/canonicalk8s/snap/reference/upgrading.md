# Upgrade notes

## Upgrade Instructions

### 1.32 -> 1.33
Updating the {{product}} snap to version 1.33 is straightforward in most cases. 
If you are not using dual stack networking, you can simply run:

```bash
sudo snap refresh k8s --channel=1.33/stable
```
All components, including Cilium, will be updated automatically.

#### Additional Steps for Dual Stack Environments
If your cluster is configured with dual stack networking (IPv4 and IPv6), 
you’ll need to make a manual adjustment before refreshing. {{product}} 1.33 
includes Cilium v1.17, which introduces a stricter requirement for dual stack: 
Each node must report both IPv4 and IPv6 addresses to the API server. 
If this is not satisfied, the Cilium agent pods will fail to start. 
For each node in the cluster, perform the following steps:
- Edit the kubelet configuration:
```bash
sudo nano /var/snap/k8s/common/args/kubelet
```
- Locate the --node-ip flag.
- Add both the IPv4 and IPv6 addresses (comma-separated) from the same interface:
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
