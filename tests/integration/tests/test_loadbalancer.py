#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
from enum import Enum
from pathlib import Path
from typing import List

import pytest
from test_util import harness, tags, util
from test_util.config import MANIFESTS_DIR, SUBSTRATE

LOG = logging.getLogger(__name__)


class K8sNetType(Enum):
    ipv4 = "ipv4"
    ipv6 = "ipv6"
    dualstack = "dualstack"


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.disable_k8s_bootstrapping()
# For loadbalancer communication
@pytest.mark.required_ports(80)
def test_loadbalancer_ipv4(instances: List[harness.Instance]):
    _test_loadbalancer(instances, k8s_net_type=K8sNetType.ipv4)


@pytest.mark.node_count(2)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.skipif(
    SUBSTRATE == "multipass", reason="QUEMU does not properly support IPv6"
)
def test_loadbalancer_ipv6_only(instances: List[harness.Instance]):
    _test_loadbalancer(instances, k8s_net_type=K8sNetType.ipv6)


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.PULL_REQUEST)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.dualstack()
@pytest.mark.network_type("dualstack")
@pytest.mark.skipif(
    SUBSTRATE == "multipass", reason="QUEMU does not properly support IPv6"
)
def test_loadbalancer_ipv6_dualstack(instances: List[harness.Instance]):
    _test_loadbalancer(instances, k8s_net_type=K8sNetType.dualstack)


def _test_loadbalancer(instances: List[harness.Instance], k8s_net_type: K8sNetType):
    instance = instances[0]
    tester_instance = instances[1]

    if k8s_net_type == K8sNetType.ipv6:
        bootstrap_config = (MANIFESTS_DIR / "bootstrap-ipv6-only.yaml").read_text()
        instance.exec(
            ["k8s", "bootstrap", "--file", "-", "--address", "::/0"],
            input=str.encode(bootstrap_config),
        )
    elif k8s_net_type == K8sNetType.dualstack:
        bootstrap_config = (MANIFESTS_DIR / "bootstrap-dualstack.yaml").read_text()
        instance.exec(
            ["k8s", "bootstrap", "--file", "-"],
            input=str.encode(bootstrap_config),
        )
    else:
        instance.exec(["k8s", "bootstrap"])

    lb_cidrs = []

    def get_lb_cidr(ipv6_cidr: bool):
        instance_default_ip = util.get_default_ip(instance, ipv6=ipv6_cidr)
        tester_instance_default_ip = util.get_default_ip(
            tester_instance, ipv6=ipv6_cidr
        )
        instance_default_cidr = util.get_default_cidr(instance, instance_default_ip)
        lb_cidr = util.find_suitable_cidr(
            parent_cidr=instance_default_cidr,
            excluded_ips=[instance_default_ip, tester_instance_default_ip],
        )
        return lb_cidr

    if k8s_net_type in (K8sNetType.ipv4, K8sNetType.dualstack):
        lb_cidrs.append(get_lb_cidr(ipv6_cidr=False))
    if k8s_net_type in (K8sNetType.ipv6, K8sNetType.dualstack):
        lb_cidrs.append(get_lb_cidr(ipv6_cidr=True))
    lb_cidr_str = ",".join(lb_cidrs)

    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)
    instance.exec(
        [
            "k8s",
            "set",
            f"load-balancer.cidrs={lb_cidr_str}",
            "load-balancer.l2-mode=true",
        ]
    )

    manifest = MANIFESTS_DIR / "loadbalancer-test.yaml"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for nginx pod to show up...")
    util.stubbornly(retries=5, delay_s=10).on(instance).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Nginx pod showed up.")

    util.stubbornly(retries=3, delay_s=5).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "run=my-nginx",
            "--timeout",
            "180s",
        ]
    )

    util.stubbornly(retries=10, delay_s=5).on(instance).until(
        lambda p: "my-nginx" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "service", "-o", "json"])

    p = (
        util.stubbornly(retries=10, delay_s=5)
        .on(instance)
        .until(lambda p: len(p.stdout.decode().replace("'", "")) > 0)
        .exec(
            [
                "k8s",
                "kubectl",
                "get",
                "service",
                "my-nginx",
                "-o=jsonpath='{.status.loadBalancer.ingress[0].ip}'",
            ],
        )
    )
    service_ip = p.stdout.decode().replace("'", "")
    if ":" in service_ip:
        service_ip = "[" + service_ip + "]"

    LOG.info(f"Reaching out to service with service_ip = {service_ip}")
    util.stubbornly(retries=40, delay_s=5).on(tester_instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", service_ip])


# ---------------------------------------------------------------------------
# BGP integration tests (CR-level; no real BGP router required)
# ---------------------------------------------------------------------------

# These tests only assert that k8sd creates the correct MetalLB CRs —
# actual BGP connectivity requires real router infra and is out of scope.


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.disable_k8s_bootstrapping()
def test_loadbalancer_bgp_single_peer(instances: List[harness.Instance]):
    """Single-peer BGP typed keys create one BGPPeer CR with correct fields.

    Uses the original single-peer typed-key path (bgp-peer-address,
    bgp-peer-asn, bgp-peer-port) and verifies that k8sd creates exactly one
    BGPPeer CR with the expected spec.  This is a regression test for the
    pre-existing single-peer BGP feature.
    """
    instance = instances[0]
    instance.exec(["k8s", "bootstrap"])
    util.wait_for_network(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)

    instance.exec(
        [
            "k8s",
            "set",
            "load-balancer.bgp-mode=true",
            "load-balancer.bgp-local-asn=65000",
            "load-balancer.bgp-peer-address=192.0.2.1",
            "load-balancer.bgp-peer-asn=65001",
            "load-balancer.bgp-peer-port=179",
            "load-balancer.cidrs=192.0.2.0/24",
        ]
    )

    LOG.info("Waiting for one BGPPeer CR to appear ...")
    p = (
        util.stubbornly(retries=20, delay_s=5)
        .on(instance)
        .until(lambda p: len(json.loads(p.stdout.decode()).get("items", [])) == 1)
        .exec(
            ["k8s", "kubectl", "get", "bgppeers", "-n", "metallb-system", "-o", "json"]
        )
    )

    peers = json.loads(p.stdout.decode())["items"]
    assert len(peers) == 1, f"expected 1 BGPPeer CR, got {len(peers)}"

    spec = peers[0]["spec"]
    assert spec["peerAddress"] == "192.0.2.1", f"peerAddress mismatch: {spec}"
    assert spec["peerASN"] == 65001, f"peerASN mismatch: {spec}"
    assert spec["myASN"] == 65000, f"myASN mismatch: {spec}"
    assert spec.get("peerPort", 179) == 179, f"peerPort mismatch: {spec}"

    LOG.info("Single BGPPeer CR has the expected fields.")


# The bgp-peers and advertise-all-pools annotations are alpha. These tests
# only assert that k8sd creates the correct BGPPeer / BGPAdvertisement CRs —
# actual BGP connectivity requires real router infra and is out of scope.
_BGP_PEERS_ANNOTATION = "k8sd/v1alpha1/metallb/bgp-peers"
_ADVERTISE_ALL_POOLS_ANNOTATION = "k8sd/v1alpha1/metallb/advertise-all-pools"

# Multi-peer YAML matching the Banco d'Italia topology (three zones).
_MULTI_PEER_ANNOTATION_VALUE = """\
- peerAddress: 192.0.2.1
  peerASN: 65001
  myASN: 65000
  nodeSelector:
    topology.kubernetes.io/zone: i1
- peerAddress: 192.0.2.2
  peerASN: 65002
  myASN: 65000
  nodeSelector:
    topology.kubernetes.io/zone: i2
- peerAddress: 192.0.2.3
  peerASN: 65003
  myASN: 65000
  nodeSelector:
    topology.kubernetes.io/zone: i3
"""


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.disable_k8s_bootstrapping()
def test_loadbalancer_bgp_multi_peer_annotation(instances: List[harness.Instance]):
    """BGP multi-peer annotation creates the correct BGPPeer CRs.

    Sets the bgp-peers annotation on a BGP-mode load balancer and verifies that
    k8sd creates exactly three BGPPeer CRs with the expected peerAddress,
    peerASN, myASN, peerPort, and nodeSelectors fields.  No real BGP router is
    required because we only inspect the Kubernetes CRs, not session state.
    """
    instance = instances[0]
    instance.exec(["k8s", "bootstrap"])
    util.wait_for_network(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)

    # Set bgp-mode and peers annotation together so validation sees the
    # annotation in the same update.
    instance.exec(
        [
            "k8s",
            "set",
            "load-balancer.bgp-mode=true",
            "load-balancer.bgp-local-asn=65000",
            "load-balancer.cidrs=192.0.2.0/24",
            f"annotations={_BGP_PEERS_ANNOTATION}: |\n"
            + "\n".join(
                f"  {line}" for line in _MULTI_PEER_ANNOTATION_VALUE.splitlines()
            )
            + "\n",
        ]
    )

    LOG.info("Waiting for three BGPPeer CRs to appear ...")
    p = (
        util.stubbornly(retries=20, delay_s=5)
        .on(instance)
        .until(lambda p: len(json.loads(p.stdout.decode()).get("items", [])) == 3)
        .exec(
            ["k8s", "kubectl", "get", "bgppeers", "-n", "metallb-system", "-o", "json"]
        )
    )

    peers = json.loads(p.stdout.decode())["items"]
    assert len(peers) == 3, f"expected 3 BGPPeer CRs, got {len(peers)}"

    # Build an index by peerAddress for deterministic assertions.
    by_addr = {peer["spec"]["peerAddress"]: peer["spec"] for peer in peers}

    expected = [
        ("192.0.2.1", 65001, "i1"),
        ("192.0.2.2", 65002, "i2"),
        ("192.0.2.3", 65003, "i3"),
    ]
    for addr, asn, zone in expected:
        assert addr in by_addr, f"BGPPeer for {addr} not found, got {list(by_addr)}"
        spec = by_addr[addr]
        assert spec["peerASN"] == asn, f"{addr}: peerASN mismatch"
        assert spec["myASN"] == 65000, f"{addr}: myASN mismatch"
        assert spec.get("peerPort", 179) == 179, f"{addr}: peerPort mismatch"
        ns = spec.get("nodeSelectors", [])
        assert ns, f"{addr}: missing nodeSelectors"
        labels = ns[0].get("matchLabels", {})
        assert (
            labels.get("topology.kubernetes.io/zone") == zone
        ), f"{addr}: nodeSelector zone mismatch (got {labels})"

    LOG.info("All three BGPPeer CRs have the expected fields.")


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.disable_k8s_bootstrapping()
def test_loadbalancer_bgp_advertise_all_pools_annotation(
    instances: List[harness.Instance],
):
    """advertise-all-pools=true produces a BGPAdvertisement with no ipAddressPools.

    Verifies that setting the advertise-all-pools annotation causes k8sd to
    render a BGPAdvertisement whose spec does NOT contain an ipAddressPools
    key — equivalent to MetalLB's empty-spec / spec:null behaviour which
    advertises all IP address pools.
    """
    instance = instances[0]
    instance.exec(["k8s", "bootstrap"])
    util.wait_for_network(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)

    instance.exec(
        [
            "k8s",
            "set",
            "load-balancer.bgp-mode=true",
            "load-balancer.bgp-local-asn=65000",
            "load-balancer.bgp-peer-address=192.0.2.1",
            "load-balancer.bgp-peer-asn=65001",
            "load-balancer.cidrs=192.0.2.0/24",
        ]
    )

    # Reconciliation is asynchronous — wait for the BGPAdvertisement to appear
    # with an ipAddressPools restriction before asserting the default behaviour.
    LOG.info("Waiting for BGPAdvertisement with ipAddressPools to appear ...")

    def _has_pool_restriction(p):
        data = json.loads(p.stdout.decode())
        items = data.get("items", [])
        if not items:
            return False
        spec = items[0].get("spec") or {}
        return "ipAddressPools" in spec

    util.stubbornly(retries=20, delay_s=5).on(instance).until(
        _has_pool_restriction
    ).exec(
        [
            "k8s",
            "kubectl",
            "get",
            "bgpadvertisements",
            "-n",
            "metallb-system",
            "-o",
            "json",
        ]
    )

    LOG.info("Default BGPAdvertisement correctly restricts to a named pool.")

    # Enable advertise-all-pools.
    instance.exec(
        [
            "k8s",
            "set",
            f'annotations={_ADVERTISE_ALL_POOLS_ANNOTATION}: "true"',
        ]
    )

    LOG.info("Waiting for BGPAdvertisement to drop ipAddressPools ...")

    def _no_pool_restriction(p):
        data = json.loads(p.stdout.decode())
        items = data.get("items", [])
        if not items:
            return False
        spec = items[0].get("spec") or {}
        return "ipAddressPools" not in spec

    util.stubbornly(retries=20, delay_s=5).on(instance).until(
        _no_pool_restriction
    ).exec(
        [
            "k8s",
            "kubectl",
            "get",
            "bgpadvertisements",
            "-n",
            "metallb-system",
            "-o",
            "json",
        ]
    )

    LOG.info("BGPAdvertisement correctly has no ipAddressPools restriction.")


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.disable_k8s_bootstrapping()
def test_loadbalancer_bgp_annotation_peers_with_advertise_all_pools(
    instances: List[harness.Instance],
):
    """Annotation BGP peers combined with advertise-all-pools=true.

    Verifies that when both bgp-peers and advertise-all-pools annotations are
    set, k8sd creates the correct BGPPeer CRs from the annotation AND renders
    a BGPAdvertisement with no ipAddressPools restriction.
    """
    instance = instances[0]
    instance.exec(["k8s", "bootstrap"])
    util.wait_for_network(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)

    instance.exec(
        [
            "k8s",
            "set",
            "load-balancer.bgp-mode=true",
            "load-balancer.bgp-local-asn=65000",
            "load-balancer.cidrs=192.0.2.0/24",
            f"annotations={_BGP_PEERS_ANNOTATION}: |\n"
            + "\n".join(
                f"  {line}" for line in _MULTI_PEER_ANNOTATION_VALUE.splitlines()
            )
            + f'\n{_ADVERTISE_ALL_POOLS_ANNOTATION}: "true"\n',
        ]
    )

    LOG.info("Waiting for 3 BGPPeer CRs with no ipAddressPools restriction ...")

    def _annotation_peers_with_advertise_all(p):
        data = json.loads(p.stdout.decode())
        items = data.get("items", [])
        return len(items) == 3

    util.stubbornly(retries=20, delay_s=5).on(instance).until(
        _annotation_peers_with_advertise_all
    ).exec(["k8s", "kubectl", "get", "bgppeers", "-n", "metallb-system", "-o", "json"])

    def _no_pool_restriction(p):
        data = json.loads(p.stdout.decode())
        items = data.get("items", [])
        if not items:
            return False
        spec = items[0].get("spec") or {}
        return "ipAddressPools" not in spec

    util.stubbornly(retries=20, delay_s=5).on(instance).until(
        _no_pool_restriction
    ).exec(
        [
            "k8s",
            "kubectl",
            "get",
            "bgpadvertisements",
            "-n",
            "metallb-system",
            "-o",
            "json",
        ]
    )

    LOG.info("BGPPeer CRs from annotation and no ipAddressPools restriction confirmed.")
