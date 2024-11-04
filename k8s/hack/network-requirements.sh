#!/usr/bin/env bash

set -u

if [ "$EUID" -ne 0 ]
then echo "Please run this script as root."
  exit 1
fi

# Require cgroup2 to be mounted
cgroup_hostroot="$(mount -t cgroup2 | head -1 | cut -d' ' -f3)"
if [ -z "$cgroup_hostroot" ]; then
  echo "cgroup2 mount not found, fail"
  exit 1
fi

# Require bpf to be mounted
# TODO: Move into init after https://bugs.launchpad.net/snapd/+bug/2048506
bpf_root="$(mount -t bpf | head -1 | cut -d' ' -f3)"
if [ -z "$bpf_root" ]; then
  if ! mount -t bpf -o rw,nosuid,nodev,noexec,relatime,mode=700 bpf /sys/fs/bpf; then
    echo "/sys/fs/bpf not found and couldn't auto mount, failing..."
    exit 1
  fi
fi

# TODO: Move into init after https://bugs.launchpad.net/snapd/+bug/2048507
# Cilium expects rp_filter=0 which is set to rp_filter=1 by defaut by systemd(Ubuntu 22.04>)
# Check https://github.com/cilium/cilium/issues/20125 for more.
sudo sed -i -e '/net.ipv4.conf.*.rp_filter/d' $(sudo grep -ril '\.rp_filter' /etc/sysctl.d/ /usr/lib/sysctl.d/)
sudo sysctl -a | grep '\.rp_filter' | awk '{print $1" = 0"}' | sudo tee -a /etc/sysctl.d/1000-cilium.conf
sudo sysctl --system

# TODO: Remove after https://bugs.launchpad.net/snapd/+bug/2047798
#sudo sed -i -e 's/\/run\/netns\/ r/\/run\/netns\/ rk/g' /var/lib/snapd/apparmor/profiles/snap.k8s.containerd
#sudo sed -i -e 's/userns,/userns rwk,/g' /var/lib/snapd/apparmor/profiles/snap.k8s.containerd # Tempfix
#sudo apparmor_parser -r /var/lib/snapd/apparmor/profiles/snap.k8s.containerd

# TODO: Remove after https://bugs.launchpad.net/snapd/+bug/2053271
sudo sed -i -e '/mountinfo/aowner @{PROC}\/@{pid}\/task\/@{tid}\/mountinfo r,' /var/lib/snapd/apparmor/profiles/snap.k8s.k8sd
sudo apparmor_parser -r /var/lib/snapd/apparmor/profiles/snap.k8s.k8sd
