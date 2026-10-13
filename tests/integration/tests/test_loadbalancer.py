#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
import textwrap
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, List, Sequence

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


# ---------------------------------------------------------------------------
# frr-k8s backend and BFD integration tests
# ---------------------------------------------------------------------------

_BGP_BACKEND_ANNOTATION = "k8sd/v1alpha1/metallb/bgp-backend"
_BFD_PROFILES_ANNOTATION = "k8sd/v1alpha1/metallb/bfd-profiles"

_METALLB_NAMESPACE = "metallb-system"
_FRR_K8S_POD_SELECTOR = "app.kubernetes.io/component=frr-k8s"
# The frr-k8s subchart must run the Canonical rocks, never the upstream images.
_FRR_K8S_IMAGE_PREFIXES = ("ghcr.io/canonical/frr-k8s:", "ghcr.io/canonical/frr:")

_STATUS_CMD = ["k8s", "status", "--output-format", "json"]
_FRR_NODE_STATES_CMD = ["k8s", "kubectl", "get", "frrnodestates", "-o", "json"]
_FRR_K8S_STATUS_MSG = "enabled, BGP mode (alpha), frr-k8s backend"
_NATIVE_STATUS_MSG = "enabled, BGP mode (alpha)"
_DISABLED_STATUS_MSG = "disabled"

_NODE_ASN = 65000
_ROUTER_ASN = 65001
_BFD_PORT = 3784

# Every BFDProfile field, all non-default: FRR omits default values from its
# running config, and each field must be mapped into the CR spec by the chart.
_FAST_FAILOVER_SPEC = {
    "receiveInterval": 150,
    "transmitInterval": 200,
    "detectMultiplier": 5,
    "echoInterval": 60,
    "echoMode": True,
    "passiveMode": True,
    "minimumTtl": 250,
}

_FAST_FAILOVER_PROFILE = """\
- name: fast-failover
  receiveInterval: 150
  transmitInterval: 200
  detectMultiplier: 5
  echoInterval: 60
  echoMode: true
  passiveMode: true
  minimumTtl: 250
"""

# FRR running config lines rendered from _FAST_FAILOVER_PROFILE.
_FAST_FAILOVER_FRR_CONFIG = [
    "profile fast-failover",
    "receive-interval 150",
    "transmit-interval 200",
    "detect-multiplier 5",
    "echo-mode",
    "echo transmit-interval 60",
    "passive-mode",
    "minimum-ttl 250",
]

# One peer uses BFD, one uses a non-default port, and one selects no node so
# it must not reach FRR at all. All share the local ASN: in FRR mode, MetalLB
# requires peers on the same node to use the same myASN.
_BFD_PEERS = """\
- peerAddress: 192.0.2.1
  peerASN: 65001
  bfdProfile: fast-failover
- peerAddress: 192.0.2.2
  peerASN: 65002
  peerPort: 1179
- peerAddress: 192.0.2.3
  peerASN: 65003
  nodeSelector:
    k8sd-test/no-such-label: "true"
"""

# The same peers without BFD, so that switching between the two updates the
# BGPPeers instead of deleting them.
_NO_BFD_PEERS = _BFD_PEERS.replace("  bfdProfile: fast-failover\n", "")

_BFD_PEERS_FRR_CONFIG = [
    "neighbor 192.0.2.1 bfd profile fast-failover",
    "neighbor 192.0.2.2 remote-as 65002",
    "neighbor 192.0.2.2 port 1179",
]
_NO_FRR_CONFIG = ["neighbor 192.0.2.2 bfd", "neighbor 192.0.2.3"]


def _annotations_arg(annotations: Dict[str, str]) -> str:
    """Render annotations as a `k8s set annotations=...` argument.

    Multi-line values become YAML block literals; a value of "-" removes the
    annotation.
    """
    rendered = ""
    for key, value in annotations.items():
        if "\n" in value:
            rendered += f"{key}: |\n" + textwrap.indent(value, "  ")
        else:
            rendered += f"{key}: {json.dumps(value)}\n"
    return f"annotations={rendered}"


def _set_bgp(instance: harness.Instance, annotations: Dict[str, str]):
    """Enable BGP mode on the load balancer with the given annotations."""
    instance.exec(
        [
            "k8s",
            "set",
            "load-balancer.bgp-mode=true",
            f"load-balancer.bgp-local-asn={_NODE_ASN}",
            "load-balancer.cidrs=192.0.2.0/24",
            _annotations_arg(annotations),
        ]
    )


def _wait_for_json(
    instance: harness.Instance,
    command: List[str],
    condition: Callable[[Any], bool],
    retries: int = 30,
    delay_s: int = 10,
) -> Any:
    """Re-run command until condition holds on its JSON output, and return it."""
    p = (
        util.stubbornly(retries=retries, delay_s=delay_s)
        .on(instance)
        .until(lambda p: condition(json.loads(p.stdout.decode())))
        .exec(command)
    )
    return json.loads(p.stdout.decode())


def _kubectl_get(kind: str, *args: str) -> List[str]:
    return [
        "k8s",
        "kubectl",
        "get",
        kind,
        *args,
        "-n",
        _METALLB_NAMESPACE,
        "-o",
        "json",
    ]


def _lb_message(status: Dict[str, Any]) -> str:
    return status.get("load-balancer", {}).get("message", "")


def _frr_config_matches(
    expected: Sequence[str], unexpected: Sequence[str] = ()
) -> Callable[[Any], bool]:
    """Check FRRNodeState's FRR running config for expected/unexpected lines."""

    def check(node_states: Dict[str, Any]) -> bool:
        items = node_states.get("items", [])
        if not items:
            return False
        config = items[0].get("status", {}).get("runningConfig", "")
        return all(line in config for line in expected) and not any(
            line in config for line in unexpected
        )

    return check


def _wait_for_valid_metallb_config(instance: harness.Instance):
    """Wait until MetalLB's controller and speakers all accept the config.

    k8s status cannot show this: MetalLB can reject a configuration k8sd
    applied successfully, e.g. a leftover BFDProfile on the native backend.
    """
    _wait_for_json(
        instance,
        _kubectl_get("configurationstates"),
        lambda d: bool(d["items"])
        and all(s.get("status", {}).get("result") == "Valid" for s in d["items"]),
    )


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.disable_k8s_bootstrapping()
def test_loadbalancer_bgp_frr_k8s_bfd_lifecycle(instances: List[harness.Instance]):
    """frr-k8s backend and BFD profiles through their whole lifecycle.

    Covers invalid backend/BFD annotations, enabling the frr-k8s backend
    (Canonical rocks, status message), BFDProfiles created from the
    bfd-profiles annotation, their wiring into BGPPeers and FRR's running
    config together with other peer fields, updating and removing profiles,
    disabling BFD, switching to the native backend and back, and disabling
    the load balancer with BFD configured. The peers are unreachable, so no
    session comes up; FRRNodeState's running config shows what FRR runs.
    """
    instance = instances[0]
    instance.exec(["k8s", "bootstrap"])
    util.wait_for_network(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)

    # Each case sets all three annotations so the previous one does not leak
    # into it, and is rejected before anything is applied.
    invalid_cases = [
        (
            {
                _BGP_BACKEND_ANNOTATION: "frr",
                _BGP_PEERS_ANNOTATION: _NO_BFD_PEERS,
                _BFD_PROFILES_ANNOTATION: "-",
            },
            'invalid bgp-backend annotation "frr"',
        ),
        (
            {
                _BGP_BACKEND_ANNOTATION: "native",
                _BGP_PEERS_ANNOTATION: _BFD_PEERS,
                _BFD_PROFILES_ANNOTATION: "-",
            },
            'bfdProfile "fast-failover" requires',
        ),
        (
            {
                _BGP_BACKEND_ANNOTATION: "native",
                _BGP_PEERS_ANNOTATION: _NO_BFD_PEERS,
                _BFD_PROFILES_ANNOTATION: _FAST_FAILOVER_PROFILE,
            },
            f"{_BFD_PROFILES_ANNOTATION} requires",
        ),
        (
            {
                _BGP_BACKEND_ANNOTATION: "frr-k8s",
                _BGP_PEERS_ANNOTATION: _BFD_PEERS,
                _BFD_PROFILES_ANNOTATION: "- name: fast-failover\n  detectMultiplier: 1\n",
            },
            "detectMultiplier 1 out of range [2, 255]",
        ),
    ]
    for annotations, error in invalid_cases:
        _set_bgp(instance, annotations)
        LOG.info("Waiting for load-balancer status to report %r ...", error)
        status = _wait_for_json(
            instance, _STATUS_CMD, lambda s: error in _lb_message(s)
        )
        assert not status["load-balancer"]["enabled"], f"not degraded: {status}"

    _set_bgp(
        instance,
        {
            _BGP_BACKEND_ANNOTATION: "frr-k8s",
            _BGP_PEERS_ANNOTATION: _BFD_PEERS,
            _BFD_PROFILES_ANNOTATION: _FAST_FAILOVER_PROFILE + "- name: spare\n",
        },
    )

    LOG.info("Waiting for the frr-k8s backend to be reported ...")
    _wait_for_json(
        instance, _STATUS_CMD, lambda s: _lb_message(s) == _FRR_K8S_STATUS_MSG
    )

    images = instance.exec(
        ["k8s", "list-images"], capture_output=True, text=True
    ).stdout.splitlines()
    for prefix in _FRR_K8S_IMAGE_PREFIXES:
        assert any(
            image.startswith(prefix) for image in images
        ), f"no {prefix} image in k8s list-images: {images}"

    # frr-k8s pods pull their images on first enable, so allow a generous wait.
    LOG.info("Waiting for the frr-k8s pods to be ready ...")
    util.stubbornly(retries=30, delay_s=10).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            _FRR_K8S_POD_SELECTOR,
            "-n",
            _METALLB_NAMESPACE,
            "--timeout=60s",
        ]
    )
    pods = json.loads(
        instance.exec(
            _kubectl_get("pods", "-l", _FRR_K8S_POD_SELECTOR), capture_output=True
        ).stdout.decode()
    )
    pod_images = {
        container["image"]
        for pod in pods["items"]
        for container in pod["spec"]["containers"]
        + pod["spec"].get("initContainers", [])
    }
    assert all(
        image.startswith(_FRR_K8S_IMAGE_PREFIXES) for image in pod_images
    ), f"frr-k8s pods use non-Canonical images: {pod_images}"

    LOG.info("Waiting for both BFDProfile CRs ...")
    profiles = _wait_for_json(
        instance, _kubectl_get("bfdprofiles"), lambda d: len(d["items"]) == 2
    )
    specs = {p["metadata"]["name"]: p.get("spec") or {} for p in profiles["items"]}
    assert specs == {
        "fast-failover": _FAST_FAILOVER_SPEC,
        "spare": {},
    }, f"BFDProfile specs mismatch: {specs}"

    LOG.info("Waiting for the BGPPeer CRs ...")
    peers = _wait_for_json(
        instance, _kubectl_get("bgppeers"), lambda d: len(d["items"]) == 3
    )
    bfd_by_peer = {
        p["spec"]["peerAddress"]: p["spec"].get("bfdProfile") for p in peers["items"]
    }
    assert bfd_by_peer == {
        "192.0.2.1": "fast-failover",
        "192.0.2.2": None,
        "192.0.2.3": None,
    }, f"BGPPeer bfdProfile mismatch: {bfd_by_peer}"

    LOG.info("Waiting for FRR running config to include the BFD wiring ...")
    _wait_for_json(
        instance,
        _FRR_NODE_STATES_CMD,
        _frr_config_matches(
            expected=_FAST_FAILOVER_FRR_CONFIG
            + _BFD_PEERS_FRR_CONFIG
            + ["profile spare"],
            unexpected=_NO_FRR_CONFIG,
        ),
        retries=60,
    )
    _wait_for_valid_metallb_config(instance)

    # Update a field and drop the unreferenced profile.
    _set_bgp(
        instance,
        {
            _BFD_PROFILES_ANNOTATION: _FAST_FAILOVER_PROFILE.replace(
                "detectMultiplier: 5", "detectMultiplier: 7"
            )
        },
    )

    LOG.info("Waiting for the updated BFDProfile and the spare one to be gone ...")
    _wait_for_json(
        instance,
        _kubectl_get("bfdprofiles"),
        lambda d: [
            (p["metadata"]["name"], p["spec"].get("detectMultiplier"))
            for p in d["items"]
        ]
        == [("fast-failover", 7)],
    )
    _wait_for_json(
        instance,
        _FRR_NODE_STATES_CMD,
        _frr_config_matches(
            expected=["detect-multiplier 7"], unexpected=["profile spare"]
        ),
    )

    # Disable BFD: the peers stay (they are updated, not deleted), so the
    # profile they referenced can be deleted in the same change.
    _set_bgp(
        instance,
        {_BGP_PEERS_ANNOTATION: _NO_BFD_PEERS, _BFD_PROFILES_ANNOTATION: "-"},
    )

    LOG.info("Waiting for BFD to be removed from the peers and FRR ...")
    _wait_for_json(instance, _kubectl_get("bfdprofiles"), lambda d: not d["items"])
    _wait_for_json(
        instance,
        _kubectl_get("bgppeers"),
        lambda d: len(d["items"]) == 3
        and not any(p["spec"].get("bfdProfile") for p in d["items"]),
    )
    _wait_for_json(
        instance,
        _FRR_NODE_STATES_CMD,
        _frr_config_matches(
            expected=["neighbor 192.0.2.1 remote-as 65001"],
            unexpected=["profile fast-failover", "neighbor 192.0.2.1 bfd"],
        ),
    )

    _set_bgp(instance, {_BGP_BACKEND_ANNOTATION: "native"})

    LOG.info("Waiting for the native backend and the frr-k8s pods to be gone ...")
    _wait_for_json(
        instance, _STATUS_CMD, lambda s: _lb_message(s) == _NATIVE_STATUS_MSG
    )
    _wait_for_json(
        instance,
        _kubectl_get("pods", "-l", _FRR_K8S_POD_SELECTOR),
        lambda d: not d["items"],
    )
    _wait_for_json(instance, _kubectl_get("bgppeers"), lambda d: len(d["items"]) == 3)
    _wait_for_valid_metallb_config(instance)

    # Switch back from native to frr-k8s with BFD in a single change.
    _set_bgp(
        instance,
        {
            _BGP_BACKEND_ANNOTATION: "frr-k8s",
            _BGP_PEERS_ANNOTATION: _BFD_PEERS,
            _BFD_PROFILES_ANNOTATION: _FAST_FAILOVER_PROFILE,
        },
    )

    LOG.info("Waiting for frr-k8s and BFD to be back ...")
    _wait_for_json(
        instance, _STATUS_CMD, lambda s: _lb_message(s) == _FRR_K8S_STATUS_MSG
    )
    _wait_for_json(
        instance,
        _FRR_NODE_STATES_CMD,
        _frr_config_matches(
            expected=_FAST_FAILOVER_FRR_CONFIG + _BFD_PEERS_FRR_CONFIG,
            unexpected=_NO_FRR_CONFIG,
        ),
        retries=60,
    )
    _wait_for_valid_metallb_config(instance)

    # MetalLB rejects deleting the BFDProfile while the BGPPeer using it
    # exists, so the first attempt may fail; the next reconcile finishes it.
    instance.exec(["k8s", "disable", "load-balancer"])

    LOG.info("Waiting for the load balancer to be disabled with BFD configured ...")
    _wait_for_json(
        instance,
        _STATUS_CMD,
        lambda s: _lb_message(s) == _DISABLED_STATUS_MSG,
        retries=60,
    )
    _wait_for_json(
        instance,
        ["k8s", "kubectl", "get", "crds", "-o", "json"],
        lambda d: not any(
            crd["metadata"]["name"].endswith(".metallb.io") for crd in d["items"]
        ),
    )

    LOG.info("frr-k8s and BFD lifecycle verified.")


def _setup_frr_router(router: harness.Instance, router_ip: str, node_ip: str):
    """Run FRR on router as an eBGP peer of node_ip with BFD enabled."""
    # Fresh images may still hold the dpkg lock, and mirrors can be flaky.
    apt = [
        "env",
        "DEBIAN_FRONTEND=noninteractive",
        "apt-get",
        "-o",
        "DPkg::Lock::Timeout=600",
    ]
    util.stubbornly(retries=5, delay_s=10).on(router).exec(apt + ["update"])
    util.stubbornly(retries=5, delay_s=10).on(router).exec(
        apt + ["install", "-y", "frr", "iptables"]
    )
    router.exec(
        [
            "sed",
            "-i",
            "-e",
            "s/^bgpd=no/bgpd=yes/",
            "-e",
            "s/^bfdd=no/bfdd=yes/",
            "/etc/frr/daemons",
        ]
    )
    router.exec(["systemctl", "restart", "frr"])

    # vtysh fails until the daemons are up after the restart.
    util.stubbornly(retries=10, delay_s=3).on(router).exec(
        [
            "vtysh",
            "-c",
            "configure terminal",
            "-c",
            f"router bgp {_ROUTER_ASN}",
            "-c",
            f"bgp router-id {router_ip}",
            # Accept routes from the node without an explicit route-map.
            "-c",
            "no bgp ebgp-requires-policy",
            "-c",
            f"neighbor {node_ip} remote-as {_NODE_ASN}",
            "-c",
            f"neighbor {node_ip} bfd",
        ]
    )


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.disable_k8s_bootstrapping()
# HTTP to the service, BGP, and BFD control packets.
@pytest.mark.required_ports(80, 179, f"{_BFD_PORT}/udp")
def test_loadbalancer_bgp_bfd_session(instances: List[harness.Instance]):
    """BGP and BFD sessions against a real FRR router, end to end.

    The second instance runs FRR as the BGP peer. Verifies that the BGP
    session establishes, that BFD comes up with the timers of the BFDProfile
    declared in the annotation, that frr-k8s reports both sessions in
    BGPSessionState, that the router learns the LoadBalancer IP over BGP and
    reaches the service through it, that losing BFD tears the BGP session
    down well before the hold timer, and that the service stays announced
    after BFD is removed and the backend switched to native.
    """
    instance, router = instances
    instance.exec(["k8s", "bootstrap"])
    node_ip = util.get_default_ip(instance)
    router_ip = util.get_default_ip(router)

    _setup_frr_router(router, router_ip, node_ip)

    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)

    peers = f"""\
- peerAddress: {router_ip}
  peerASN: {_ROUTER_ASN}
  bfdProfile: fast-failover
"""
    _set_bgp(
        instance,
        {
            _BGP_BACKEND_ANNOTATION: "frr-k8s",
            _BGP_PEERS_ANNOTATION: peers,
            _BFD_PROFILES_ANNOTATION: """\
- name: fast-failover
  receiveInterval: 150
  transmitInterval: 200
  detectMultiplier: 5
""",
        },
    )

    neighbor_cmd = ["vtysh", "-c", f"show bgp neighbors {node_ip} json"]
    bfd_peers_cmd = ["vtysh", "-c", "show bfd peers json"]

    def _established(d: Dict[str, Any]) -> bool:
        return d.get(node_ip, {}).get("bgpState") == "Established"

    def _bfd_peer(peers: List[Dict[str, Any]]) -> Dict[str, Any]:
        return next((p for p in peers if p.get("peer") == node_ip), {})

    # Right after coming up, BFD still runs its slow start timers, so wait for
    # the negotiated profile timers rather than the first "up".
    def _bfd_ready(peers: List[Dict[str, Any]]) -> bool:
        p = _bfd_peer(peers)
        return p.get("status") == "up" and (
            p.get("remote-receive-interval"),
            p.get("remote-transmit-interval"),
            p.get("remote-detect-multiplier"),
        ) == (150, 200, 5)

    # frr-k8s pods pull their images on first enable, so allow a generous wait.
    LOG.info("Waiting for the BGP session to be established ...")
    _wait_for_json(router, neighbor_cmd, _established, retries=60)

    LOG.info("Waiting for the BFD session to come up with the profile timers ...")
    _wait_for_json(router, bfd_peers_cmd, _bfd_ready)

    # frr-status polls FRR every 2 minutes.
    LOG.info("Waiting for frr-k8s to report both sessions ...")
    _wait_for_json(
        instance,
        ["k8s", "kubectl", "get", "bgpsessionstates", "-A", "-o", "json"],
        lambda d: any(
            s.get("status", {}).get("peer") == router_ip
            and s["status"].get("bgpStatus") == "Established"
            and s["status"].get("bfdStatus") == "Up"
            for s in d["items"]
        ),
    )

    manifest = MANIFESTS_DIR / "loadbalancer-test.yaml"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )
    service = _wait_for_json(
        instance,
        ["k8s", "kubectl", "get", "service", "my-nginx", "-o", "json"],
        lambda d: d.get("status", {}).get("loadBalancer", {}).get("ingress"),
        retries=20,
        delay_s=5,
    )
    service_ip = service["status"]["loadBalancer"]["ingress"][0]["ip"]

    def _route_learned(d: Dict[str, Any]) -> bool:
        return f"{service_ip}/32" in d.get("routes", {})

    def _reach_service():
        util.stubbornly(retries=40, delay_s=5).on(router).until(
            lambda p: "Welcome to nginx!" in p.stdout.decode()
        ).exec(["curl", "--max-time", "5", service_ip])

    LOG.info("Waiting for the router to learn %s over BGP ...", service_ip)
    _wait_for_json(router, ["vtysh", "-c", "show ip bgp json"], _route_learned)
    _reach_service()

    # Drop BFD control packets on the router: BFD must detect the failure
    # within its ~1.5s detection time and tear BGP down, long before the BGP
    # hold timer (90s) would.
    drop_bfd = ["INPUT", "-p", "udp", "--dport", str(_BFD_PORT), "-j", "DROP"]
    router.exec(["iptables", "-I", *drop_bfd])

    LOG.info("Waiting for BFD to bring the BGP session down ...")
    _wait_for_json(
        router,
        bfd_peers_cmd,
        lambda d: _bfd_peer(d).get("status") == "down",
        retries=10,
        delay_s=2,
    )
    _wait_for_json(
        router,
        neighbor_cmd,
        lambda d: d.get(node_ip, {}).get("lastResetDueTo") == "BFD down received",
        retries=10,
        delay_s=2,
    )

    router.exec(["iptables", "-D", *drop_bfd])

    LOG.info("Waiting for BGP and BFD to recover ...")
    _wait_for_json(router, neighbor_cmd, _established)
    _wait_for_json(router, bfd_peers_cmd, _bfd_ready)

    # Stop using BFD on both ends first, then switch to the native backend
    # with the session live.
    router.exec(
        [
            "vtysh",
            "-c",
            "configure terminal",
            "-c",
            f"router bgp {_ROUTER_ASN}",
            "-c",
            f"no neighbor {node_ip} bfd",
        ]
    )
    _set_bgp(
        instance,
        {
            _BGP_PEERS_ANNOTATION: peers.replace("  bfdProfile: fast-failover\n", ""),
            _BFD_PROFILES_ANNOTATION: "-",
        },
    )
    _wait_for_json(instance, _kubectl_get("bfdprofiles"), lambda d: not d["items"])
    _set_bgp(instance, {_BGP_BACKEND_ANNOTATION: "native"})

    LOG.info("Waiting for the native backend to take over the session ...")
    _wait_for_json(
        instance, _STATUS_CMD, lambda s: _lb_message(s) == _NATIVE_STATUS_MSG
    )
    _wait_for_json(
        instance,
        _kubectl_get("pods", "-l", _FRR_K8S_POD_SELECTOR),
        lambda d: not d["items"],
    )
    _wait_for_valid_metallb_config(instance)
    _wait_for_json(router, neighbor_cmd, _established)
    _wait_for_json(router, ["vtysh", "-c", "show ip bgp json"], _route_learned)
    _reach_service()

    LOG.info("BGP and BFD sessions, failover and backend switch verified.")
