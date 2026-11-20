#
# Copyright 2026 Canonical, Ltd.
#
from unittest.mock import MagicMock, patch

import pytest
from test_util import tags, util


@pytest.mark.tags(tags.PULL_REQUEST)
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


@pytest.mark.tags(tags.PULL_REQUEST)
@patch("test_util.util.collect_host_apparmor_denials")
@patch("test_util.util.collect_instance_apparmor_denials")
def test_assert_no_apparmor_denials_respects_baseline_and_fails_on_new(
    mock_inst_denials, mock_host_denials
):
    inst = MagicMock()
    inst.id = "node-1"
    old_line = 'type=AVC msg=audit(1.0:1): apparmor="DENIED" profile="snap.k8s.k8sd" name="/old"'
    new_line = 'type=AVC msg=audit(2.0:2): apparmor="DENIED" profile="snap.k8s.containerd" name="/new"'

    mock_inst_denials.return_value = []
    mock_host_denials.return_value = [old_line]
    util.assert_no_apparmor_denials([inst], baseline={old_line})

    mock_host_denials.return_value = [old_line, new_line]
    with pytest.raises(AssertionError, match=r"\[host\].*snap\.k8s\.containerd"):
        util.assert_no_apparmor_denials([inst], baseline={old_line})
