#
# Copyright 2026 Canonical, Ltd.
#
import ipaddress
import subprocess
from unittest.mock import MagicMock, patch

import pytest
from test_util import tags, util


@pytest.mark.tags(tags.PULL_REQUEST)
def test_find_suitable_cidr_ipv4_default():
    """Verify that find_suitable_cidr picks the first valid /30 above network address (.0)."""
    cidr = util.find_suitable_cidr("10.145.252.0/24")
    assert cidr == "10.145.252.4/30"


@pytest.mark.tags(tags.PULL_REQUEST)
def test_find_suitable_cidr_ipv4_excludes_specified_ips():
    """Verify that candidate blocks containing excluded IPs (like .4 registry mirror) are skipped."""
    cidr = util.find_suitable_cidr("10.145.252.0/24", excluded_ips=["10.145.252.4"])
    assert cidr == "10.145.252.8/30"

    cidr2 = util.find_suitable_cidr(
        "10.145.252.0/24", excluded_ips=["10.145.252.5", "10.145.252.9"]
    )
    assert cidr2 == "10.145.252.12/30"


@pytest.mark.tags(tags.PULL_REQUEST)
def test_find_suitable_cidr_ipv6_basic_and_excluded():
    """Verify that IPv6 picks /126 blocks and avoids excluded IPv6 addresses."""
    cidr = util.find_suitable_cidr("fd00::/64")
    assert cidr == "fd00::4/126"

    cidr_excluded = util.find_suitable_cidr("fd00::/64", excluded_ips=["fd00::4"])
    assert cidr_excluded == "fd00::8/126"

    # IPv6 has no broadcast address, so /125 can select the second /126 block
    cidr_narrow = util.find_suitable_cidr("fd00::/125")
    assert cidr_narrow == "fd00::4/126"


@pytest.mark.tags(tags.PULL_REQUEST)
def test_find_suitable_cidr_ipv6_does_not_exclude_highest_address():
    """Verify IPv6 selection does not treat the highest address as an IPv4 broadcast."""
    assert util.find_suitable_cidr("fd00::/125") == "fd00::4/126"


@pytest.mark.tags(tags.PULL_REQUEST)
def test_find_suitable_cidr_skips_network_and_broadcast():
    """Verify that network (.0) and broadcast (.15 in /28) are never allocated."""
    # Usable /30 blocks in 10.0.0.0/28 are .4 and .8; .0 and .12 hold network and broadcast.
    cidr = util.find_suitable_cidr("10.0.0.0/28")
    assert cidr == "10.0.0.4/30"

    cidr = util.find_suitable_cidr("10.0.0.0/28", excluded_ips=["10.0.0.4"])
    assert cidr == "10.0.0.8/30"

    # No usable /30 left.
    with pytest.raises(RuntimeError):
        util.find_suitable_cidr("10.0.0.0/28", excluded_ips=["10.0.0.4", "10.0.0.8"])


@pytest.mark.tags(tags.PULL_REQUEST)
def test_find_suitable_cidr_too_small_raises():
    """Verify that parent CIDR smaller than /30 raises RuntimeError."""
    with pytest.raises(RuntimeError):
        util.find_suitable_cidr("10.0.0.0/30")

    with pytest.raises(RuntimeError):
        util.find_suitable_cidr("10.0.0.0/31")


@pytest.mark.tags(tags.PULL_REQUEST)
def test_find_suitable_cidr_discovers_harness_instances():
    """Verify that harness instance IPs (including registry mirror) are discovered and excluded."""
    mock_instance = MagicMock()
    mock_harness = MagicMock()
    mock_harness.instances = {"node-1", "k8s-integration-1-registry"}
    mock_instance._h = mock_harness
    mock_instance.id = "node-1"
    mock_instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=""
    )

    def mock_get_default_ip(inst, ipv6=False):
        if (
            getattr(inst, "id", None) == "k8s-integration-1-registry"
            or getattr(inst, "_id", None) == "k8s-integration-1-registry"
        ):
            return "10.145.252.4"
        return "10.145.252.132"

    with patch("test_util.util.get_default_ip", side_effect=mock_get_default_ip):
        with patch("test_util.util._is_candidate_occupied", return_value=False):
            cidr = util.find_suitable_cidr("10.145.252.0/24", instance=mock_instance)
            assert cidr == "10.145.252.8/30"


@pytest.mark.tags(tags.PULL_REQUEST)
def test_find_suitable_cidr_discovers_gateway_and_neighbor():
    """Verify that gateway IP and existing neighbor entries on the instance are excluded."""
    mock_instance = MagicMock()
    mock_instance._h = None

    def mock_exec(cmd, **kwargs):
        cmd_str = " ".join(cmd)
        if "route" in cmd_str and "show" in cmd_str:
            return subprocess.CompletedProcess(
                args=cmd,
                returncode=0,
                stdout="default via 10.145.252.1 dev eth0 proto dhcp metric 100",
            )
        if "neigh" in cmd_str and "show" in cmd_str:
            return subprocess.CompletedProcess(
                args=cmd,
                returncode=0,
                stdout="10.145.252.4 dev eth0 lladdr 00:16:3e:ab:cd:ef REACHABLE\n",
            )
        return subprocess.CompletedProcess(args=cmd, returncode=0, stdout="")

    mock_instance.exec.side_effect = mock_exec

    with patch("test_util.util._is_candidate_occupied", return_value=False):
        cidr = util.find_suitable_cidr("10.145.252.0/24", instance=mock_instance)
        assert cidr == "10.145.252.8/30"


@pytest.mark.tags(tags.PULL_REQUEST)
def test_find_suitable_cidr_active_probe_skips_occupied_block():
    """Verify that active probing detects occupied candidate blocks and advances to next."""
    mock_instance = MagicMock()
    mock_instance._h = None
    mock_instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=""
    )

    def mock_probe(instance, candidate):
        # Pretend 10.145.252.4/30 is occupied on the wire
        if str(candidate) == "10.145.252.4/30":
            return True
        return False

    with patch("test_util.util._is_candidate_occupied", side_effect=mock_probe):
        cidr = util.find_suitable_cidr("10.145.252.0/24", instance=mock_instance)
        assert cidr == "10.145.252.8/30"


@pytest.mark.tags(tags.PULL_REQUEST)
def test_is_candidate_occupied_probe_results():
    """Verify _is_candidate_occupied parses probe script results accurately."""
    mock_instance = MagicMock()
    net = ipaddress.ip_network("10.145.252.4/30")

    mock_instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=b"occupied 10.145.252.4\n"
    )
    assert util._is_candidate_occupied(mock_instance, net) is True

    mock_instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=b"free\n"
    )
    assert util._is_candidate_occupied(mock_instance, net) is False

    # Inconclusive probe output (empty or malformed) must not be accepted as free.
    mock_instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=1, stdout=b""
    )
    with pytest.raises(RuntimeError):
        util._is_candidate_occupied(mock_instance, net)

    mock_instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=1, stdout=b"free\n"
    )
    with pytest.raises(RuntimeError):
        util._is_candidate_occupied(mock_instance, net)

    mock_instance.exec.return_value = subprocess.CompletedProcess(
        args=[], returncode=0, stdout=b"not free\n"
    )
    with pytest.raises(RuntimeError):
        util._is_candidate_occupied(mock_instance, net)

    # Probe failure/exception must not be accepted as free.
    mock_instance.exec.side_effect = RuntimeError("exec failed")
    with pytest.raises(RuntimeError):
        util._is_candidate_occupied(mock_instance, net)
