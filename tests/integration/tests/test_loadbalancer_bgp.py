#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
from string import Template
from typing import Any, Dict, List, Sequence

import pytest
import yaml
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


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
BGP_PEERS_ANNOTATION = "k8sd/v1alpha1/metallb/bgp-peers"
ADVERTISE_ALL_POOLS_ANNOTATION = "k8sd/v1alpha1/metallb/advertise-all-pools"

# Multi-peer YAML matching the Banco d'Italia topology (three zones).
MULTI_PEER_ANNOTATION_VALUE = """\
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
            f"annotations={BGP_PEERS_ANNOTATION}: |\n"
            + "\n".join(
                f"  {line}" for line in MULTI_PEER_ANNOTATION_VALUE.splitlines()
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
            f'annotations={ADVERTISE_ALL_POOLS_ANNOTATION}: "true"',
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
            f"annotations={BGP_PEERS_ANNOTATION}: |\n"
            + "\n".join(
                f"  {line}" for line in MULTI_PEER_ANNOTATION_VALUE.splitlines()
            )
            + f'\n{ADVERTISE_ALL_POOLS_ANNOTATION}: "true"\n',
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

BGP_BACKEND_ANNOTATION = "k8sd/v1alpha1/metallb/bgp-backend"
BFD_PROFILES_ANNOTATION = "k8sd/v1alpha1/metallb/bfd-profiles"

METALLB_NAMESPACE = "metallb-system"
FRR_K8S_POD_SELECTOR = "app.kubernetes.io/component=frr-k8s"
# The frr-k8s subchart must run the Canonical rocks, never the upstream images.
FRR_K8S_IMAGE_PREFIXES = ("ghcr.io/canonical/frr-k8s:", "ghcr.io/canonical/frr:")

STATUS_CMD = ["k8s", "status", "--output-format", "json"]
FRR_NODE_STATES_CMD = ["k8s", "kubectl", "get", "frrnodestates", "-o", "json"]
FRR_K8S_STATUS = "enabled, BGP mode (alpha), frr-k8s backend"
NATIVE_STATUS = "enabled, BGP mode (alpha)"
DISABLED_STATUS = "disabled"

NODE_ASN = 65000
ROUTER_ASN = 65001
BFD_PORT = 3784

# Annotations k8sd must reject, with the error k8s status reports for each.
INVALID_BFD_CONFIGS = {
    "invalid-backend.yaml": 'invalid bgp-backend annotation "frr"',
    "bfd-peer-on-native.yaml": 'bfdProfile "fast-failover" requires',
    "bfd-profiles-on-native.yaml": f"{BFD_PROFILES_ANNOTATION} requires",
    "bfd-profile-out-of-range.yaml": "detectMultiplier 1 out of range [2, 255]",
}

# FRR running config lines rendered from frr-k8s-bfd.yaml.
FRR_K8S_BFD_CONFIG = [
    "profile fast-failover",
    "receive-interval 150",
    "transmit-interval 200",
    "detect-multiplier 5",
    "echo-mode",
    "echo transmit-interval 60",
    "passive-mode",
    "minimum-ttl 250",
    "profile spare",
    "neighbor 192.0.2.1 bfd profile fast-failover",
    "neighbor 192.0.2.2 remote-as 65002",
    "neighbor 192.0.2.2 port 1179",
]
# FRR running config lines frr-k8s-bfd.yaml must not render.
FRR_K8S_BFD_ABSENT_CONFIG = ["neighbor 192.0.2.2 bfd", "neighbor 192.0.2.3"]


def load_annotations(name: str, **substitutes: str) -> str:
    """Read a BGP template's annotations, filling in its $ placeholders."""
    return Template((config.BGP_DIR / name).read_text()).substitute(substitutes)


def bfd_profiles(annotations: str) -> Dict[str, Dict[str, Any]]:
    """Map the name of each profile in the bfd-profiles annotation to its fields."""
    profiles = yaml.safe_load(yaml.safe_load(annotations)[BFD_PROFILES_ANNOTATION])
    return {p.pop("name"): p for p in profiles}


def set_bgp(instance: harness.Instance, annotations: str):
    """Enable BGP mode on the load balancer with the given annotations."""
    instance.exec(
        [
            "k8s",
            "set",
            "load-balancer.bgp-mode=true",
            f"load-balancer.bgp-local-asn={NODE_ASN}",
            "load-balancer.cidrs=192.0.2.0/24",
            f"annotations={annotations}",
        ]
    )


def metallb_get(kind: str, *args: str) -> List[str]:
    """kubectl command that gets the kind resources in MetalLB's namespace."""
    return ["k8s", "kubectl", "get", kind, *args, "-n", METALLB_NAMESPACE, "-o", "json"]


def lb_message(status: Dict[str, Any]) -> str:
    return status.get("load-balancer", {}).get("message", "")


def wait_for_lb_status(instance: harness.Instance, message: str):
    """Wait until k8s status reports message for the load balancer."""
    LOG.info("Waiting for load-balancer status %r ...", message)
    util.wait_for_json(
        instance, STATUS_CMD, lambda s: lb_message(s) == message, retries=60
    )


def wait_for_lb_error(instance: harness.Instance, error: str):
    """Wait until k8s status reports the load balancer degraded with error."""
    LOG.info("Waiting for load-balancer status to report %r ...", error)
    status = util.wait_for_json(instance, STATUS_CMD, lambda s: error in lb_message(s))
    assert not status["load-balancer"]["enabled"], f"not degraded: {status}"


def wait_for_frr_config(
    instance: harness.Instance,
    expected: Sequence[str],
    unexpected: Sequence[str] = (),
    retries: int = 30,
):
    """Wait until FRR runs every expected config line and no unexpected one.

    The peers are unreachable in most tests, so no session comes up;
    FRRNodeState's running config shows what FRR runs regardless.
    """

    def matches(node_states: Dict[str, Any]) -> bool:
        items = node_states.get("items", [])
        if not items:
            return False

        running = items[0].get("status", {}).get("runningConfig", "")
        return all(line in running for line in expected) and not any(
            line in running for line in unexpected
        )

    util.wait_for_json(instance, FRR_NODE_STATES_CMD, matches, retries=retries)


def wait_for_valid_metallb_config(instance: harness.Instance):
    """Wait until MetalLB's controller and speakers all accept the config.

    k8s status cannot show this: MetalLB can reject a configuration k8sd
    applied successfully, e.g. a leftover BFDProfile on the native backend.
    """
    util.wait_for_json(
        instance,
        metallb_get("configurationstates"),
        lambda d: bool(d["items"])
        and all(s.get("status", {}).get("result") == "Valid" for s in d["items"]),
    )


def wait_for_frr_k8s_pods(instance: harness.Instance):
    """Wait for the frr-k8s pods and check they run the Canonical rocks."""
    images = instance.exec(
        ["k8s", "list-images"], capture_output=True, text=True
    ).stdout.splitlines()
    for prefix in FRR_K8S_IMAGE_PREFIXES:
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
            FRR_K8S_POD_SELECTOR,
            "-n",
            METALLB_NAMESPACE,
            "--timeout=60s",
        ]
    )

    pods = json.loads(
        instance.exec(
            metallb_get("pods", "-l", FRR_K8S_POD_SELECTOR), capture_output=True
        ).stdout.decode()
    )
    pod_images = {
        container["image"]
        for pod in pods["items"]
        for container in pod["spec"]["containers"]
        + pod["spec"].get("initContainers", [])
    }
    assert all(
        image.startswith(FRR_K8S_IMAGE_PREFIXES) for image in pod_images
    ), f"frr-k8s pods use non-Canonical images: {pod_images}"


def wait_for_native_backend(instance: harness.Instance):
    """Wait for the native backend to be reported and the frr-k8s pods gone."""
    wait_for_lb_status(instance, NATIVE_STATUS)
    LOG.info("Waiting for the frr-k8s pods to be gone ...")
    util.wait_for_json(
        instance,
        metallb_get("pods", "-l", FRR_K8S_POD_SELECTOR),
        lambda d: not d["items"],
    )
    wait_for_valid_metallb_config(instance)


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
    the load balancer with BFD configured.
    """
    instance = instances[0]
    instance.exec(["k8s", "bootstrap"])
    util.wait_for_network(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)

    # Each invalid config sets all three annotations so the previous one does
    # not leak into it, and is rejected before anything is applied.
    for name, error in INVALID_BFD_CONFIGS.items():
        set_bgp(instance, load_annotations(name))
        wait_for_lb_error(instance, error)

    frr_k8s_bfd = load_annotations("frr-k8s-bfd.yaml")
    set_bgp(instance, frr_k8s_bfd)
    wait_for_lb_status(instance, FRR_K8S_STATUS)
    wait_for_frr_k8s_pods(instance)

    LOG.info("Waiting for the BFDProfile CRs ...")
    profiles = util.wait_for_json(
        instance, metallb_get("bfdprofiles"), lambda d: len(d["items"]) == 2
    )
    specs = {p["metadata"]["name"]: p.get("spec") or {} for p in profiles["items"]}
    assert specs == bfd_profiles(frr_k8s_bfd), f"BFDProfile specs mismatch: {specs}"

    LOG.info("Waiting for the BGPPeer CRs ...")
    peers = util.wait_for_json(
        instance, metallb_get("bgppeers"), lambda d: len(d["items"]) == 3
    )
    bfd_by_peer = {
        p["spec"]["peerAddress"]: p["spec"].get("bfdProfile") for p in peers["items"]
    }
    assert bfd_by_peer == {
        "192.0.2.1": "fast-failover",
        "192.0.2.2": None,
        "192.0.2.3": None,
    }, f"BGPPeer bfdProfile mismatch: {bfd_by_peer}"

    LOG.info("Waiting for FRR to run the BFD config ...")
    wait_for_frr_config(
        instance, FRR_K8S_BFD_CONFIG, FRR_K8S_BFD_ABSENT_CONFIG, retries=60
    )
    wait_for_valid_metallb_config(instance)

    set_bgp(instance, load_annotations("bfd-profile-updated.yaml"))

    LOG.info("Waiting for the updated BFDProfile and the spare one to be gone ...")
    util.wait_for_json(
        instance,
        metallb_get("bfdprofiles"),
        lambda d: [
            (p["metadata"]["name"], p["spec"].get("detectMultiplier"))
            for p in d["items"]
        ]
        == [("fast-failover", 7)],
    )
    wait_for_frr_config(
        instance, expected=["detect-multiplier 7"], unexpected=["profile spare"]
    )

    set_bgp(instance, load_annotations("bfd-disabled.yaml"))

    LOG.info("Waiting for BFD to be removed from the peers and FRR ...")
    util.wait_for_json(instance, metallb_get("bfdprofiles"), lambda d: not d["items"])
    util.wait_for_json(
        instance,
        metallb_get("bgppeers"),
        lambda d: len(d["items"]) == 3
        and not any(p["spec"].get("bfdProfile") for p in d["items"]),
    )
    # The unused "profile fast-failover" block stays in FRR: MetalLB's speaker
    # skips syncing an empty set of BFD profiles, so the last one is not
    # removed from the FRRConfiguration (metallb v0.16.1).
    wait_for_frr_config(
        instance,
        expected=["neighbor 192.0.2.1 remote-as 65001"],
        unexpected=["neighbor 192.0.2.1 bfd"],
    )

    set_bgp(instance, f"{BGP_BACKEND_ANNOTATION}: native")
    wait_for_native_backend(instance)
    util.wait_for_json(
        instance, metallb_get("bgppeers"), lambda d: len(d["items"]) == 3
    )

    # Switch back from native to frr-k8s with BFD in a single change.
    set_bgp(instance, frr_k8s_bfd)
    wait_for_lb_status(instance, FRR_K8S_STATUS)
    wait_for_frr_config(
        instance, FRR_K8S_BFD_CONFIG, FRR_K8S_BFD_ABSENT_CONFIG, retries=60
    )
    wait_for_valid_metallb_config(instance)

    # MetalLB rejects deleting the BFDProfile while the BGPPeer using it
    # exists, so the first attempt may fail; the next reconcile finishes it.
    instance.exec(["k8s", "disable", "load-balancer"])
    wait_for_lb_status(instance, DISABLED_STATUS)
    LOG.info("Waiting for the MetalLB CRDs to be gone ...")
    util.wait_for_json(
        instance,
        ["k8s", "kubectl", "get", "crds", "-o", "json"],
        lambda d: not any(
            crd["metadata"]["name"].endswith(".metallb.io") for crd in d["items"]
        ),
    )

    LOG.info("frr-k8s and BFD lifecycle verified.")


def configure_router(router: harness.Instance, *commands: str):
    """Run commands in the BGP router section of the FRR router's config."""
    args = ["configure terminal", f"router bgp {ROUTER_ASN}", *commands]
    # vtysh fails until the daemons are up after a restart.
    util.stubbornly(retries=10, delay_s=3).on(router).exec(
        ["vtysh", *(arg for command in args for arg in ("-c", command))]
    )


def setup_frr_router(router: harness.Instance, router_ip: str, node_ip: str):
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

    configure_router(
        router,
        f"bgp router-id {router_ip}",
        # Accept routes from the node without an explicit route-map.
        "no bgp ebgp-requires-policy",
        f"neighbor {node_ip} remote-as {NODE_ASN}",
        f"neighbor {node_ip} bfd",
    )


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.disable_k8s_bootstrapping()
# HTTP to the service, BGP, and BFD control packets.
@pytest.mark.required_ports(80, 179, f"{BFD_PORT}/udp")
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

    setup_frr_router(router, router_ip, node_ip)

    util.wait_for_network(instance)
    util.wait_for_dns(instance)

    instance.exec(["k8s", "enable", "load-balancer"])
    util.wait_for_load_balancer(instance)

    router_peer = {"ROUTER_IP": router_ip, "ROUTER_ASN": str(ROUTER_ASN)}
    bfd_session = load_annotations("bfd-session.yaml", **router_peer)
    set_bgp(instance, bfd_session)

    profile = bfd_profiles(bfd_session)["fast-failover"]
    profile_timers = (
        profile["receiveInterval"],
        profile["transmitInterval"],
        profile["detectMultiplier"],
    )

    neighbor_cmd = ["vtysh", "-c", f"show bgp neighbors {node_ip} json"]
    bfd_peers_cmd = ["vtysh", "-c", "show bfd peers json"]
    routes_cmd = ["vtysh", "-c", "show ip bgp json"]

    def established(neighbors: Dict[str, Any]) -> bool:
        return neighbors.get(node_ip, {}).get("bgpState") == "Established"

    def connections_dropped(neighbors: Dict[str, Any]) -> int:
        return neighbors.get(node_ip, {}).get("connectionsDropped", 0)

    def bfd_peer(peers: List[Dict[str, Any]]) -> Dict[str, Any]:
        return next((p for p in peers if p.get("peer") == node_ip), {})

    # Right after coming up, BFD still runs its slow start timers, so wait for
    # the negotiated profile timers rather than the first "up".
    def bfd_ready(peers: List[Dict[str, Any]]) -> bool:
        peer = bfd_peer(peers)
        return (
            peer.get("status") == "up"
            and (
                peer.get("remote-receive-interval"),
                peer.get("remote-transmit-interval"),
                peer.get("remote-detect-multiplier"),
            )
            == profile_timers
        )

    # frr-k8s pods pull their images on first enable, so allow a generous wait.
    LOG.info("Waiting for the BGP session to be established ...")
    util.wait_for_json(router, neighbor_cmd, established, retries=60)

    LOG.info("Waiting for the BFD session to come up with the profile timers ...")
    util.wait_for_json(router, bfd_peers_cmd, bfd_ready)

    # frr-status polls FRR every 2 minutes.
    LOG.info("Waiting for frr-k8s to report both sessions ...")
    util.wait_for_json(
        instance,
        ["k8s", "kubectl", "get", "bgpsessionstates", "-A", "-o", "json"],
        lambda d: any(
            s.get("status", {}).get("peer") == router_ip
            and s["status"].get("bgpStatus") == "Established"
            and s["status"].get("bfdStatus") == "Up"
            for s in d["items"]
        ),
    )

    manifest = config.MANIFESTS_DIR / "loadbalancer-test.yaml"
    instance.exec(["k8s", "kubectl", "apply", "-f", "-"], input=manifest.read_bytes())
    service = util.wait_for_json(
        instance,
        ["k8s", "kubectl", "get", "service", "my-nginx", "-o", "json"],
        lambda d: d.get("status", {}).get("loadBalancer", {}).get("ingress"),
        retries=20,
        delay_s=5,
    )
    service_ip = service["status"]["loadBalancer"]["ingress"][0]["ip"]

    def route_learned(routes: Dict[str, Any]) -> bool:
        return f"{service_ip}/32" in routes.get("routes", {})

    def reach_service():
        util.stubbornly(retries=40, delay_s=5).on(router).until(
            lambda p: "Welcome to nginx!" in p.stdout.decode()
        ).exec(["curl", "--max-time", "5", service_ip])

    LOG.info("Waiting for the router to learn %s over BGP ...", service_ip)
    util.wait_for_json(router, routes_cmd, route_learned)
    reach_service()

    dropped = connections_dropped(
        json.loads(router.exec(neighbor_cmd, capture_output=True).stdout.decode())
    )

    # Drop BFD control packets on the router: BFD must detect the failure
    # within its ~1.5s detection time and tear BGP down, long before the BGP
    # hold timer (90s) would. TCP is not blocked, so BGP reconnects right away
    # and overwrites lastResetDueTo; count the dropped sessions instead.
    drop_bfd = ["INPUT", "-p", "udp", "--dport", str(BFD_PORT), "-j", "DROP"]
    router.exec(["iptables", "-I", *drop_bfd])

    LOG.info("Waiting for BFD to bring the BGP session down ...")
    util.wait_for_json(
        router,
        bfd_peers_cmd,
        lambda d: bfd_peer(d).get("status") == "down",
        retries=10,
        delay_s=2,
    )
    util.wait_for_json(
        router,
        neighbor_cmd,
        lambda d: connections_dropped(d) > dropped,
        retries=10,
        delay_s=2,
    )

    router.exec(["iptables", "-D", *drop_bfd])

    LOG.info("Waiting for BGP and BFD to recover ...")
    util.wait_for_json(router, neighbor_cmd, established)
    util.wait_for_json(router, bfd_peers_cmd, bfd_ready)

    # Stop using BFD on both ends first, then switch to the native backend
    # with the session live.
    configure_router(router, f"no neighbor {node_ip} bfd")
    set_bgp(instance, load_annotations("bfd-session-disabled.yaml", **router_peer))
    util.wait_for_json(instance, metallb_get("bfdprofiles"), lambda d: not d["items"])

    set_bgp(instance, f"{BGP_BACKEND_ANNOTATION}: native")
    wait_for_native_backend(instance)

    LOG.info("Waiting for the native backend to take over the session ...")
    util.wait_for_json(router, neighbor_cmd, established)
    util.wait_for_json(router, routes_cmd, route_learned)
    reach_service()

    LOG.info("BGP and BFD sessions, failover and backend switch verified.")
