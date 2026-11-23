#
# Copyright 2026 Canonical, Ltd.
#
from test_util import (  # noqa: F401 (harness first avoids a circular import)
    harness,
    util,
)


def test_filter_k8s_apparmor_denials():
    k8sd_line = (
        'audit: type=1400 audit(1.1:1): apparmor="DENIED" '
        'operation="open" profile="snap.k8s.k8sd" name="/run/snapd.socket"'
    )
    lxd_peer_line = (
        'type=AVC msg=audit(1.2:2): apparmor="DENIED" '
        'operation="signal" profile=":lxd-c1:snap.k8s.containerd" '
        'peer="cri-containerd.apparmor.d"'
    )
    pod_line = (
        'audit: type=1400 audit(1.3:3): apparmor="DENIED" '
        'operation="open" profile="cri-containerd.apparmor.d" '
        'name="/proc/sysrq-trigger"'
    )
    not_mount_noise = (
        'audit: type=1400 audit(1.4:4): apparmor="DENIED" '
        'operation="not mount" profile="snap.k8s.kubelet" name="/"'
    )
    glibc_hugepages_probe = (
        'audit: type=1400 audit(1.5:5): apparmor="DENIED" '
        'operation="open" class="file" profile="snap.k8s.k8s-apiserver-proxy" '
        'name="/sys/kernel/mm/hugepages/" requested_mask="r" denied_mask="r"'
    )
    hugepages_write = (
        'audit: type=1400 audit(1.6:6): apparmor="DENIED" '
        'operation="open" class="file" profile="snap.k8s.k8s-apiserver-proxy" '
        'name="/sys/kernel/mm/hugepages/hugepages-2048kB/nr_hugepages" '
        'requested_mask="w" denied_mask="w"'
    )
    other_profile = (
        'audit: type=1400 audit(1.7:7): apparmor="DENIED" '
        'operation="open" profile="/usr/sbin/rsyslogd" name="/etc/foo"'
    )

    sample = "\n".join(
        [
            k8sd_line,
            lxd_peer_line,
            pod_line,
            not_mount_noise,
            glibc_hugepages_probe,
            hugepages_write,
            other_profile,
            k8sd_line,
        ]
    )
    assert util._filter_k8s_apparmor_denials(sample) == [
        k8sd_line,
        lxd_peer_line,
        pod_line,
        hugepages_write,
    ]
