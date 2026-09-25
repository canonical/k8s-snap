#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

CILIUM_AGENT_SELECTOR = "k8s-app=cilium"
CILIUM_OPERATOR_SELECTOR = "io.cilium/app=operator"
CILIUM_AGENT_CONTAINER = "cilium-agent"

# A node label no node carries, used to strand the Cilium agent DaemonSet.
UNSCHEDULABLE_KEY = "ck8s.io/nonexistent"


def _status_patterns(cluster_status: str) -> List[str]:
    """Expected `k8s status` output for the bootstrap-network-dns.yaml config.

    Only network and dns are enabled, so the remaining feature lines are matched
    by name alone.
    """
    return [
        rf"cluster status:\s*{cluster_status}",
        r"control plane nodes:\s*(\d{1,3}(?:\.\d{1,3}){3}:\d{1,5})\s\(voter\)",
        r"high availability:\s*no",
        r"datastore:\s*etcd",
        r"network:\s*enabled",
        r"dns:\s*enabled at (\d{1,3}(?:\.\d{1,3}){3})",
        r"ingress:",
        r"load-balancer:",
        r"local-storage:",
        r"gateway",
    ]


READY_PATTERNS = _status_patterns("ready")
NOT_READY_PATTERNS = _status_patterns("not ready")


def _pods_gone(instance: harness.Instance, selector: str):
    """Block until no pod matches `selector`.

    Deliberately not `kubectl wait --for=delete`: that exits non-zero with
    `no matching resources found` when the pod is already gone, and the agent
    DaemonSet's terminationGracePeriodSeconds is 1.
    """
    util.stubbornly(retries=60, delay_s=2).on(instance).until(
        lambda p: p.stdout.decode().strip() == ""
    ).exec(
        [
            "k8s",
            "kubectl",
            "-n",
            "kube-system",
            "get",
            "pod",
            "-l",
            selector,
            "-o",
            "name",
        ]
    )


def _assert_cluster_not_ready(instance: harness.Instance, context: str):
    """`k8s status` must report `not ready`, and must do so because of the CNI.

    The node is checked separately and must still be Ready. Kubelet reports
    NodeReady as soon as a CNI config file exists on disk, and that file stays
    behind when Cilium is removed, so the node keeps looking healthy while the
    network is broken. That is the condition from canonical/k8s-snap#1789, and it
    rules out `HasReadyNodes` as the reason readiness was withheld.
    """
    LOG.info("%s: waiting for `k8s status` to report not ready", context)
    util.stubbornly(retries=30, delay_s=2).on(instance).until(
        lambda p: util.status_output_matches(p, NOT_READY_PATTERNS)
    ).exec(["k8s", "status"])

    assert len(util.ready_nodes(instance)) == 1, (
        f"{context}: the node stopped being Ready, so this no longer reproduces "
        "canonical/k8s-snap#1789 (readiness could be withheld by HasReadyNodes alone)"
    )

    p = instance.exec(
        ["k8s", "status", "--wait-ready", "--timeout", "30s"],
        capture_output=True,
        check=False,
        text=True,
    )
    assert p.returncode != 0, (
        f"{context}: `k8s status --wait-ready` returned success while the CNI "
        f"was down (this is canonical/k8s-snap#1789). stdout={p.stdout!r} "
        f"stderr={p.stderr!r}"
    )

    p = instance.exec(
        ["k8s", "x-wait-for", "network", "--timeout", "10s"],
        capture_output=True,
        check=False,
        text=True,
    )
    assert p.returncode != 0, (
        f"{context}: the network is reported as ready while the CNI is down. "
        f"stdout={p.stdout!r} stderr={p.stderr!r}"
    )


def _assert_cluster_ready(instance: harness.Instance, context: str):
    """`k8s status --wait-ready` must succeed once the CNI is healthy."""
    LOG.info("%s: waiting for `k8s status` to report ready", context)
    util.stubbornly(retries=15, delay_s=10).on(instance).until(
        lambda p: util.status_output_matches(p, READY_PATTERNS)
    ).exec(["k8s", "status", "--wait-ready"])


def _agent_readiness_probe(instance: harness.Instance) -> dict:
    """The Cilium agent container's current readinessProbe, resolved by name."""
    ds = json.loads(
        instance.exec(
            [
                "k8s",
                "kubectl",
                "-n",
                "kube-system",
                "get",
                "daemonset",
                "cilium",
                "-o",
                "json",
            ],
            capture_output=True,
        ).stdout.decode()
    )
    container = next(
        c
        for c in ds["spec"]["template"]["spec"]["containers"]
        if c["name"] == CILIUM_AGENT_CONTAINER
    )
    return container["readinessProbe"]


def _patch_agent_readiness_probe(instance: harness.Instance, probe: dict):
    """Strategic-merge the agent container's readinessProbe, keyed by name."""
    instance.exec(
        [
            "k8s",
            "kubectl",
            "-n",
            "kube-system",
            "patch",
            "daemonset",
            "cilium",
            "-p",
            json.dumps(
                {
                    "spec": {
                        "template": {
                            "spec": {
                                "containers": [
                                    {
                                        "name": CILIUM_AGENT_CONTAINER,
                                        "readinessProbe": probe,
                                    }
                                ]
                            }
                        }
                    }
                }
            ),
        ]
    )


def _agent_running_not_ready(p) -> bool:
    """True when some agent pod is Running with Ready != True.

    A DaemonSet rolling update briefly leaves zero agent pods, and an absent pod
    is a different readiness failure, so the pod must be present and Running.
    """
    for pod in json.loads(p.stdout.decode())["items"]:
        if pod["status"].get("phase") != "Running":
            continue
        condition = next(
            (
                c
                for c in pod["status"].get("conditions", [])
                if c.get("type") == "Ready"
            ),
            None,
        )
        if condition is not None and str(condition.get("status")).lower() != "true":
            return True
    return False


def _prepare_cluster(instance: harness.Instance):
    """Bring the cluster to a healthy baseline before breaking the CNI."""
    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_network(instance)
    _assert_cluster_ready(instance, "baseline")


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-network-dns.yaml").read_text()
)
@pytest.mark.tags(tags.NIGHTLY)
def test_status_not_ready_when_cilium_operator_missing(
    instances: List[harness.Instance],
):
    """`k8s status` must not report ready while the Cilium operator is missing.

    Reproduces canonical/k8s-snap#1789 for the operator pods, which are the first
    workload the daemon's network readiness check inspects.
    """
    instance = instances[0]
    _prepare_cluster(instance)

    LOG.info("Scaling cilium-operator to zero replicas")
    instance.exec(
        [
            "k8s",
            "kubectl",
            "-n",
            "kube-system",
            "scale",
            "deployment/cilium-operator",
            "--replicas=0",
        ]
    )
    _pods_gone(instance, CILIUM_OPERATOR_SELECTOR)
    _assert_cluster_not_ready(instance, "cilium-operator scaled to zero")

    LOG.info("Restoring cilium-operator")
    instance.exec(
        [
            "k8s",
            "kubectl",
            "-n",
            "kube-system",
            "scale",
            "deployment/cilium-operator",
            "--replicas=1",
        ]
    )
    util.wait_for_pods_ready(
        instance, namespace="kube-system", label_selector=CILIUM_OPERATOR_SELECTOR
    )
    _assert_cluster_ready(instance, "cilium-operator restored")


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-network-dns.yaml").read_text()
)
@pytest.mark.tags(tags.NIGHTLY)
def test_status_not_ready_when_cilium_agent_missing(
    instances: List[harness.Instance],
):
    """`k8s status` must not report ready while no Cilium agent is scheduled.

    Reproduces canonical/k8s-snap#1789 for the agent pods. With no agent at all
    the CNI config file stays on disk, so the node keeps reporting Ready.
    """
    instance = instances[0]
    _prepare_cluster(instance)

    LOG.info("Stranding the cilium agent DaemonSet on a node selector no node matches")
    instance.exec(
        [
            "k8s",
            "kubectl",
            "-n",
            "kube-system",
            "patch",
            "daemonset",
            "cilium",
            "-p",
            json.dumps(
                {
                    "spec": {
                        "template": {
                            "spec": {"nodeSelector": {UNSCHEDULABLE_KEY: "true"}}
                        }
                    }
                }
            ),
        ]
    )
    _pods_gone(instance, CILIUM_AGENT_SELECTOR)
    _assert_cluster_not_ready(instance, "cilium agent DaemonSet stranded")

    LOG.info("Restoring the cilium agent DaemonSet")
    # A strategic-merge null deletes the injected key and is idempotent; an RFC
    # 6902 `remove` would hard-fail if the key were somehow already absent.
    instance.exec(
        [
            "k8s",
            "kubectl",
            "-n",
            "kube-system",
            "patch",
            "daemonset",
            "cilium",
            "-p",
            json.dumps(
                {
                    "spec": {
                        "template": {
                            "spec": {"nodeSelector": {UNSCHEDULABLE_KEY: None}}
                        }
                    }
                }
            ),
        ]
    )
    util.wait_for_pods_ready(
        instance, namespace="kube-system", label_selector=CILIUM_AGENT_SELECTOR
    )
    _assert_cluster_ready(instance, "cilium agent restored")


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-network-dns.yaml").read_text()
)
@pytest.mark.tags(tags.NIGHTLY)
def test_status_not_ready_when_cilium_agent_not_ready(
    instances: List[harness.Instance],
):
    """`k8s status` must not report ready while the Cilium agent is not Ready.

    This is the signature canonical/k8s-snap#1789 is titled after: a crash-looping
    agent stays in phase Running with Ready=False. Rigging the readiness probe
    produces that state while leaving the agent process, and therefore the
    datapath and the node's Ready condition, intact.
    """
    instance = instances[0]
    _prepare_cluster(instance)

    original_probe = _agent_readiness_probe(instance)

    LOG.info("Forcing the cilium agent readiness probe to fail")
    _patch_agent_readiness_probe(
        instance,
        {
            "exec": {"command": ["/bin/false"]},
            "httpGet": None,
            "tcpSocket": None,
            "grpc": None,
            "initialDelaySeconds": 0,
            "periodSeconds": 2,
            "timeoutSeconds": 1,
            "successThreshold": 1,
            "failureThreshold": 1,
        },
    )
    util.stubbornly(retries=60, delay_s=5).on(instance).until(
        _agent_running_not_ready
    ).exec(
        [
            "k8s",
            "kubectl",
            "-n",
            "kube-system",
            "get",
            "pods",
            "-l",
            CILIUM_AGENT_SELECTOR,
            "-o",
            "json",
        ]
    )
    _assert_cluster_not_ready(instance, "cilium agent Running but not Ready")

    LOG.info("Restoring the cilium agent readiness probe")
    # `exec` must be nulled out explicitly: strategic merge would otherwise keep
    # it alongside the restored handler, which the API server rejects.
    _patch_agent_readiness_probe(instance, {"exec": None, **original_probe})
    util.wait_for_pods_ready(
        instance, namespace="kube-system", label_selector=CILIUM_AGENT_SELECTOR
    )
    _assert_cluster_ready(instance, "cilium agent readiness probe restored")
