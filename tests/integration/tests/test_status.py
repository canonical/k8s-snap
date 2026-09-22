#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
import re
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

# `cilium.CheckNetwork` in k8sd selects the CNI workloads by exactly these labels.
CILIUM_AGENT_SELECTOR = "k8s-app=cilium"
CILIUM_OPERATOR_SELECTOR = "io.cilium/app=operator"
CILIUM_AGENT_CONTAINER = "cilium-agent"

# A node label no node carries, used to strand the Cilium agent DaemonSet.
UNSCHEDULABLE_KEY = "ck8s.io/nonexistent"

CLUSTER_NOT_READY = re.compile(r"cluster status:\s+not ready")
CLUSTER_READY = re.compile(r"cluster status:\s+ready")


def _journal_mark(instance: harness.Instance) -> str:
    """A `journalctl --since` timestamp in the instance's own local time."""
    return instance.exec(
        ["date", "+%Y-%m-%d %H:%M:%S"], capture_output=True, text=True
    ).stdout.strip()


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


def _assert_cluster_not_ready(
    instance: harness.Instance, context: str, since: str, expected_error: str
):
    """`k8s status` must report `not ready`, and must do so *because of the CNI*.

    Three facts are required together:

    * `k8s status` prints `not ready`;
    * the node is still Ready in the same poll iteration — that is the #1789
      condition (kubelet reports NodeReady as soon as a CNI conflist exists on
      disk) and it rules out `HasReadyNodes` as the reason readiness was withheld;
    * k8sd logged the CNI gate firing, with the specific `CheckNetwork` error for
      this phase, after `since` — positive attribution to the gate rather than to
      the sibling `--cluster-dns` gate.

    Each poll iteration logs the `cluster status:` line it observed together with
    the ready-node count, so an exhausted poll says why it failed: an unfixed
    daemon keeps reporting `cluster status: ready` (that is #1789), whereas a
    node that genuinely went NotReady shows a ready-node count of 0.
    """

    def cni_gated_not_ready(p) -> bool:
        status = p.stdout.decode()
        cluster_line = next(
            (line.strip() for line in status.splitlines() if "cluster status:" in line),
            "<no cluster status line>",
        )
        ready = len(util.ready_nodes(instance))
        LOG.info("%s: observed %r with %d ready node(s)", context, cluster_line, ready)
        return CLUSTER_NOT_READY.search(status) is not None and ready == 1

    util.stubbornly(retries=30, delay_s=2).on(instance).until(cni_gated_not_ready).exec(
        ["k8s", "status"]
    )

    assert len(util.ready_nodes(instance)) == 1, (
        f"{context}: the node stopped being Ready, so this test no longer "
        "reproduces #1789 (readiness could be withheld by HasReadyNodes alone)"
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

    journal = instance.exec(
        [
            "journalctl",
            "-u",
            "snap.k8s.k8sd",
            "--since",
            since,
            "--no-pager",
            "-o",
            "cat",
        ],
        capture_output=True,
        text=True,
    ).stdout
    assert "network pods are not ready" in journal, (
        f"{context}: readiness was withheld, but k8sd never logged the CNI gate "
        f"firing — something other than canonical/k8sd#93 caused it"
    )
    assert expected_error in journal, (
        f"{context}: the CNI gate fired, but not with the expected check "
        f"failure {expected_error!r}"
    )


def _assert_cluster_ready(instance: harness.Instance, context: str):
    """`k8s status --wait-ready` must succeed again once the CNI is restored."""
    p = instance.exec(
        ["k8s", "status", "--wait-ready", "--timeout", "5m"],
        capture_output=True,
        check=False,
        text=True,
    )
    assert p.returncode == 0 and CLUSTER_READY.search(p.stdout), (
        f"{context}: cluster did not report ready after the CNI was restored: "
        f"rc={p.returncode} stdout={p.stdout!r} stderr={p.stderr!r}"
    )


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

    Required so this phase cannot pass for the previous phase's reason: a
    DaemonSet rolling update briefly leaves zero agent pods, and an absent pod
    is a different `CheckNetwork` branch.
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


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config(
    (config.MANIFESTS_DIR / "bootstrap-network-dns.yaml").read_text()
)
@pytest.mark.tags(tags.PULL_REQUEST)
def test_status_withholds_ready_when_cni_not_ready(instances: List[harness.Instance]):
    """`k8s status` must not report the cluster ready while the CNI is down.

    Reproduces canonical/k8s-snap#1789: kubelet marks the node Ready as soon as a
    CNI config file exists on disk, which happens before the Cilium workloads are
    serving, so `k8s status --wait-ready` used to return `ready` while Cilium pods
    were missing or crash-looping. Requires the readiness gate from
    canonical/k8sd#93; against a snap built from k8sd main this test fails.

    Each phase drives a different branch of `cilium.CheckNetwork`:
      1. operator pods absent   -> `cilium-operator pods not yet ready: no pods ...`
      2. agent pods absent      -> `cilium pods not yet ready: no pods ...`
      3. agent Running, !Ready  -> `cilium pods not yet ready: pods [...] not ready`
    Phase 3 is the crash-loop signature the issue is titled after.
    """
    instance = instances[0]

    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_network(instance)
    _assert_cluster_ready(instance, "baseline")

    # --- Phase 1: the operator is gone (CheckNetwork's first selector) ---------
    LOG.info("Scaling cilium-operator to zero replicas")
    since = _journal_mark(instance)
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
    _assert_cluster_not_ready(
        instance,
        "cilium-operator scaled to zero",
        since,
        "cilium-operator pods not yet ready: no pods in kube-system namespace on the cluster",
    )

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

    # --- Phase 2: the agent is gone (CheckNetwork's second selector) -----------
    # The #1789 premise proper: with no Cilium agent at all, the CNI conflist
    # stays on disk (`cni.uninstall: false`), so the node keeps reporting Ready.
    LOG.info("Stranding the cilium agent DaemonSet on a node selector no node matches")
    since = _journal_mark(instance)
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
    _assert_cluster_not_ready(
        instance,
        "cilium agent DaemonSet stranded",
        since,
        "cilium pods not yet ready: no pods in kube-system namespace on the cluster",
    )

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

    # --- Phase 3: the agent is Running but never Ready ------------------------
    # The signature #1789 is titled after: a crash-looping agent stays in phase
    # Running with Ready=False. Rigging the readiness probe produces exactly that
    # state while leaving the agent process — and therefore the datapath and the
    # node's Ready condition — intact.
    original_probe = _agent_readiness_probe(instance)
    LOG.info("Forcing the cilium agent readiness probe to fail")
    since = _journal_mark(instance)
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
    _assert_cluster_not_ready(
        instance,
        "cilium agent Running but not Ready",
        since,
        "cilium pods not yet ready: pods [",
    )

    LOG.info("Restoring the cilium agent readiness probe")
    # `exec` must be nulled out explicitly: strategic merge would otherwise keep
    # it alongside the restored handler, which the API server rejects.
    _patch_agent_readiness_probe(instance, {"exec": None, **original_probe})
    util.wait_for_pods_ready(
        instance, namespace="kube-system", label_selector=CILIUM_AGENT_SELECTOR
    )
    _assert_cluster_ready(instance, "cilium agent readiness probe restored")
