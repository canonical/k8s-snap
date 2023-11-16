## k8s-snap

> *NOTE*: This is work in progress, please do not share externally.

Canonical Kubernetes is an opinionated and CNCF conformant Kubernetes operated by Snaps and Charms, which come together to bring simplified operations and an enhanced security posture on any infrastructure.

### Build

```bash
sudo snap install snapcraft --classic
snapcraft
sudo mv k8s_*.snap k8s.snap
```

### Install single-node

#### Setup LXD (for quick throwaway dev)

```bash
lxc profile create k8s
cat ./test/lxd-profile.yaml | lxc profile edit k8s

lxc launch ubuntu:22.04 -p default -p k8s u1
lxc shell u1
```

#### Install and initialize snap

```bash
sudo snap install ./k8s.snap --dangerous
sudo /snap/k8s/current/k8s/init.sh

# Initialize and start services for single-node (cluster is empty)
sudo k8s init
sudo k8s start

# Kubectl commands
sudo k8s kubectl get pod,node -A
```
