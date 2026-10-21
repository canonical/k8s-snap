#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
from string import Template
from typing import Any, Dict, List, Sequence

import yaml
from test_util import config, harness, util

LOG = logging.getLogger(__name__)

BGP_PEERS_ANNOTATION = "k8sd/v1alpha1/metallb/bgp-peers"
ADVERTISE_ALL_POOLS_ANNOTATION = "k8sd/v1alpha1/metallb/advertise-all-pools"
BGP_BACKEND_ANNOTATION = "k8sd/v1alpha1/metallb/bgp-backend"
BFD_PROFILES_ANNOTATION = "k8sd/v1alpha1/metallb/bfd-profiles"

METALLB_NAMESPACE = "metallb-system"
FRR_K8S_POD_SELECTOR = "app.kubernetes.io/component=frr-k8s"
# The frr-k8s subchart must run the Canonical rocks, never the upstream images.
FRR_K8S_IMAGE_PREFIXES = ("ghcr.io/canonical/frr-k8s:", "ghcr.io/canonical/frr:")

STATUS_CMD = ["k8s", "status", "--output-format", "json"]
FRR_NODE_STATES_CMD = ["k8s", "kubectl", "get", "frrnodestates", "-o", "json"]
BGP_STATUS = "enabled, BGP mode (alpha)"
DISABLED_STATUS = "disabled"

NODE_ASN = 65000
ROUTER_ASN = 65001
BFD_PORT = 3784


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


def wait_for_bgp_status(instance: harness.Instance):
    """Wait until k8s status reports the load balancer enabled in BGP mode."""
    LOG.info("Waiting for load-balancer status %r ...", BGP_STATUS)
    util.wait_for_json(
        instance,
        STATUS_CMD,
        lambda s: lb_message(s).startswith(BGP_STATUS),
        retries=60,
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
    """Wait for the frr-k8s pods to be gone and MetalLB to accept the config."""
    LOG.info("Waiting for the frr-k8s pods to be gone ...")
    util.wait_for_json(
        instance,
        metallb_get("pods", "-l", FRR_K8S_POD_SELECTOR),
        lambda d: not d["items"],
    )
    wait_for_valid_metallb_config(instance)


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
