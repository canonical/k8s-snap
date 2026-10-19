#
# Copyright 2026 Canonical, Ltd.
#
import base64
import json
import logging
import re
import subprocess
import time
from typing import Callable, List, Optional

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)

K8SD_STATE_DIR = "/var/snap/k8s/common/var/lib/k8sd/state"
ETCD_ARGS_FILE = "/var/snap/k8s/common/args/etcd"
FORCE_NEW_CLUSTER_ARG = '--force-new-cluster="true"'

# `k8s remove-node` defaults to a 90s server-side timeout and the CLI adds 30s on
# top. This bounds the subprocess in case the CLI itself hangs.
REMOVE_NODE_TIMEOUT_S = 300


def _wait_for(
    condition: Callable[[], bool], timeout_s: int, interval_s: int, what: str
):
    deadline = time.monotonic() + timeout_s
    last_error = None
    while time.monotonic() < deadline:
        try:
            if condition():
                return
            last_error = None
        except Exception as e:  # noqa: BLE001 - retried until the deadline
            last_error = e
        time.sleep(interval_s)
    raise TimeoutError(f"timed out after {timeout_s}s waiting for {what}: {last_error}")


def _etcd_api(
    instance: harness.Instance, path: str, body: Optional[dict] = None
) -> dict:
    """POST to the node's local etcd gRPC gateway with the apiserver client certificate."""
    proc = instance.exec(
        [
            "curl",
            "--silent",
            "--show-error",
            "--fail",
            "--max-time",
            "10",
            "--cacert",
            "/etc/kubernetes/pki/etcd/ca.crt",
            "--cert",
            "/etc/kubernetes/pki/apiserver-etcd-client.crt",
            "--key",
            "/etc/kubernetes/pki/apiserver-etcd-client.key",
            f"https://127.0.0.1:2379{path}",
            "-X",
            "POST",
            "-d",
            json.dumps(body or {}),
        ],
        capture_output=True,
        text=True,
    )
    return json.loads(proc.stdout)


def _etcd_members(instance: harness.Instance) -> List[dict]:
    """List the etcd members served by the local member (serializable, works without quorum)."""
    members = _etcd_api(
        instance, "/v3/cluster/member/list", {"linearizable": False}
    ).get("members", [])
    LOG.info(
        "etcd members seen from %s: %s",
        instance.id,
        [
            (m.get("name", ""), m.get("peerURLs", []), m.get("isLearner", False))
            for m in members
        ],
    )
    return members


def _etcd_started_voters(instance: harness.Instance) -> List[str]:
    # A member that has not started yet has no name.
    return [
        m["name"]
        for m in _etcd_members(instance)
        if m.get("name") and not m.get("isLearner", False)
    ]


def _etcd_put(instance: harness.Instance, key: str) -> bool:
    """Write a key through raft. Fails (raises) when etcd has no quorum."""
    _etcd_api(
        instance,
        "/v3/kv/put",
        {
            "key": base64.b64encode(key.encode()).decode(),
            "value": base64.b64encode(str(time.time()).encode()).decode(),
        },
    )
    return True


def _etcd_has_no_quorum(instance: harness.Instance) -> bool:
    try:
        _etcd_put(instance, "/integration-test/remove-force-quorum")
    except subprocess.CalledProcessError:
        return True
    return False


def _microcluster_members(instance: harness.Instance) -> List[str]:
    """List the microcluster members from dqlite only.

    `k8s status` is not used here: while etcd has no quorum, kube-apiserver is
    down too.
    """
    proc = instance.exec(
        [
            "/snap/k8s/current/bin/k8sd",
            "--state-dir",
            K8SD_STATE_DIR,
            "sql",
            "SELECT name FROM core_cluster_members",
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        timeout=60,
    )
    # The rows are printed as [[name1] [name2]].
    members = re.findall(r"\[([^\[\]\s]+)\]", proc.stdout)
    LOG.info("microcluster members seen from %s: %s", instance.id, members)
    return members


def _kubernetes_nodes(instance: harness.Instance) -> List[str]:
    return [node["metadata"]["name"] for node in util.get_nodes(instance)]


def _remove_node_force(
    instance: harness.Instance, node_name: str
) -> subprocess.CompletedProcess:
    start = time.monotonic()
    result = instance.exec(
        ["k8s", "remove-node", node_name, "--force"],
        capture_output=True,
        text=True,
        check=False,
        timeout=REMOVE_NODE_TIMEOUT_S,
    )
    LOG.info(
        "k8s remove-node %s --force exited %d after %.0fs\nstdout: %s\nstderr: %s",
        node_name,
        result.returncode,
        time.monotonic() - start,
        result.stdout.strip(),
        result.stderr.strip(),
    )
    return result


def _restart_etcd_and_wait_for_write(instance: harness.Instance, what: str):
    instance.exec(["snap", "restart", "k8s.etcd"])
    _wait_for(
        lambda: _etcd_put(instance, "/integration-test/remove-force-quorum"),
        timeout_s=180,
        interval_s=3,
        what=what,
    )


def _join_control_planes(cp1: harness.Instance, joiners: List[harness.Instance]):
    """Join control-plane nodes to cp1 and wait until all of them are etcd voters."""
    control_planes = [cp1, *joiners]
    util.wait_until_k8s_ready(cp1, [cp1])
    for joiner in joiners:
        util.join_cluster(joiner, util.get_join_token(cp1, joiner))
    util.wait_until_k8s_ready(cp1, control_planes)
    assert util.get_datastore_type(cp1) == "etcd", "test needs the etcd datastore"
    _wait_for(
        lambda: sorted(_etcd_started_voters(cp1))
        == sorted(cp.id for cp in control_planes),
        timeout_s=180,
        interval_s=5,
        what="all control-plane nodes to be started etcd voters",
    )


def _dqlite_spare(
    cp1: harness.Instance, control_planes: List[harness.Instance]
) -> harness.Instance:
    """Return the control-plane node that holds the dqlite spare role."""
    status = json.loads(
        cp1.exec(
            ["k8s", "status", "--output-format", "json"],
            capture_output=True,
            text=True,
        ).stdout
    )
    roles = {m["name"]: m["datastore-role"] for m in status["members"]}
    LOG.info("dqlite roles: %s", roles)
    spares = [cp for cp in control_planes if roles.get(cp.id) == "spare"]
    assert len(spares) == 1, f"expected exactly one dqlite spare, got roles {roles}"
    return spares[0]


@pytest.mark.node_count(2)
@pytest.mark.tags(tags.NIGHTLY)
def test_remove_force_keeps_etcd_member_without_quorum(
    instances: List[harness.Instance],
):
    """`remove-node --force` must not drop a node that is still an etcd voter.

    Regression test for canonical/k8sd#161. On a 2-node control plane, the
    dqlite spare dies permanently. etcd still has both nodes as voters, so it
    has no quorum. k8sd could not remove the dead node from etcd, but with
    --force it removed it from microcluster anyway and reported success. The
    stale etcd voter was then out of reach of every k8sd path.

    The removal must fail and keep the node in microcluster. After etcd is
    recovered with --force-new-cluster on the survivor, the same removal must
    succeed.
    """
    cp1, cp2 = instances
    _join_control_planes(cp1, [cp2])

    spare = _dqlite_spare(cp1, [cp1, cp2])
    survivor = cp2 if spare is cp1 else cp1
    spare_name = spare.id

    # The dqlite spare dies permanently. dqlite keeps its only voter, but etcd
    # has lost one of its two voters, and with it the quorum.
    LOG.info("Deleting the dqlite spare %s", spare_name)
    spare.delete()
    _wait_for(
        lambda: _etcd_has_no_quorum(survivor),
        timeout_s=120,
        interval_s=5,
        what=f"etcd on {survivor.id} to lose quorum",
    )

    # Core assertion. Before the fix, this printed "Removed ... from cluster."
    # and exited 0, while the dead node stayed an etcd voter.
    result = _remove_node_force(survivor, spare_name)
    assert result.returncode != 0, (
        "remove-node --force must fail while the node may still be an etcd member "
        "and etcd has no quorum (canonical/k8sd#161)"
    )
    assert (
        "may still be an etcd member" in result.stderr
        and "--force-new-cluster" in result.stderr
    ), f"unexpected remove-node error: {result.stderr}"
    assert spare_name in _microcluster_members(
        survivor
    ), "the node must stay in microcluster so that a retry can still remove it from etcd"
    assert spare_name in [
        m.get("name") for m in _etcd_members(survivor)
    ], "the node must still be an etcd member"

    # Recover etcd on the survivor with --force-new-cluster, then drop the flag
    # again so that later etcd restarts don't reset the membership.
    LOG.info("Recovering etcd on %s with --force-new-cluster", survivor.id)
    survivor.exec(
        ["tee", "-a", ETCD_ARGS_FILE],
        input=f"{FORCE_NEW_CLUSTER_ARG}\n",
        text=True,
        stdout=subprocess.DEVNULL,
    )
    _restart_etcd_and_wait_for_write(
        survivor, "an etcd write after --force-new-cluster"
    )
    assert [m.get("name") for m in _etcd_members(survivor)] == [
        survivor.id
    ], "--force-new-cluster must leave the survivor as the only etcd member"

    survivor.exec(["sed", "-i", "/^--force-new-cluster/d", ETCD_ARGS_FILE])
    _restart_etcd_and_wait_for_write(
        survivor, "an etcd write after removing --force-new-cluster"
    )

    # With the node gone from etcd, the retry finishes the cleanup.
    result = _remove_node_force(survivor, spare_name)
    assert result.returncode == 0, f"retry of remove-node failed: {result.stderr}"
    assert spare_name not in _microcluster_members(survivor)
    assert spare_name not in [m.get("name") for m in _etcd_members(survivor)]
    _wait_for(
        lambda: spare_name not in _kubernetes_nodes(survivor),
        timeout_s=120,
        interval_s=5,
        what=f"the Kubernetes node {spare_name} to be deleted",
    )
    util.wait_until_k8s_ready(survivor, [survivor])
    _wait_for(
        lambda: not util.diverged_cluster_memberships(survivor, [survivor]),
        timeout_s=120,
        interval_s=5,
        what=f"cluster memberships to converge on {survivor.id}",
    )


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_remove_force_rejected_by_etcd_keeps_member(
    instances: List[harness.Instance],
):
    """`remove-node --force` must not drop a node whose removal etcd rejects.

    On a 3-node control plane, cp2 dies permanently. etcd keeps its quorum
    (cp1 and cp3), but its strict reconfiguration check rejects removing the
    live cp3, because only 1 of the remaining 2 voters would be started. The
    removal must fail, keep cp3 in microcluster, and must not point at
    --force-new-cluster, which would drop the healthy cp3. Removing the dead
    cp2 first, as the error suggests, must then let cp3 be removed.
    """
    cp1, cp2, cp3 = instances
    _join_control_planes(cp1, [cp2, cp3])
    cp2_name, cp3_name = cp2.id, cp3.id

    LOG.info("Deleting %s", cp2_name)
    cp2.delete()
    _wait_for(
        lambda: _etcd_put(cp1, "/integration-test/remove-force-quorum"),
        timeout_s=120,
        interval_s=3,
        what="an etcd write on cp1 with 2 of 3 voters up",
    )
    # etcd only counts a peer as started if it was connected within the last
    # 5s (its health interval). Give it time to notice that cp2 is gone.
    time.sleep(15)

    result = _remove_node_force(cp1, cp3_name)
    assert (
        result.returncode != 0
    ), "remove-node --force must fail while etcd refuses to remove the node"
    assert (
        "etcd did not remove" in result.stderr
    ), f"unexpected remove-node error: {result.stderr}"
    assert (
        "--force-new-cluster" not in result.stderr
    ), "etcd has quorum, so the error must not suggest --force-new-cluster"
    assert cp3_name in _microcluster_members(
        cp1
    ), "the node must stay in microcluster so that a retry can still remove it from etcd"
    assert cp3_name in [m.get("name") for m in _etcd_members(cp1)]
    assert _etcd_put(cp1, "/integration-test/remove-force-quorum")

    # Remove the dead peer first, then the original removal goes through.
    result = _remove_node_force(cp1, cp2_name)
    assert result.returncode == 0, f"removal of the dead {cp2_name} failed"
    result = _remove_node_force(cp1, cp3_name)
    assert result.returncode == 0, f"retry of the removal of {cp3_name} failed"
    assert cp3_name not in _microcluster_members(cp1)
    assert [m.get("name") for m in _etcd_members(cp1)] == [cp1.id]
    util.wait_until_k8s_ready(cp1, [cp1])
    _wait_for(
        lambda: not util.diverged_cluster_memberships(cp1, [cp1]),
        timeout_s=180,
        interval_s=5,
        what="cluster memberships to converge on cp1",
    )


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_remove_force_worker_without_etcd_quorum(
    instances: List[harness.Instance],
):
    """`remove-node --force` of a dead worker stays best-effort without etcd quorum.

    A worker is never an etcd member. Whether a node is a worker is decided
    from the microcluster membership (dqlite), which still works when etcd has
    lost its quorum. So removing a dead worker must still succeed.
    """
    cp1, cp2, worker = instances
    _join_control_planes(cp1, [cp2])
    util.join_cluster(worker, util.get_join_token(cp1, worker, "--worker"))
    util.wait_until_k8s_ready(cp1, [cp1, cp2, worker])

    spare = _dqlite_spare(cp1, [cp1, cp2])
    survivor = cp2 if spare is cp1 else cp1
    worker_name = worker.id

    LOG.info("Deleting the dqlite spare %s and the worker %s", spare.id, worker_name)
    spare.delete()
    worker.delete()
    _wait_for(
        lambda: _etcd_has_no_quorum(survivor),
        timeout_s=120,
        interval_s=5,
        what=f"etcd on {survivor.id} to lose quorum",
    )

    result = _remove_node_force(survivor, worker_name)
    assert (
        result.returncode == 0
    ), f"removal of a dead worker must stay best-effort: {result.stderr}"
