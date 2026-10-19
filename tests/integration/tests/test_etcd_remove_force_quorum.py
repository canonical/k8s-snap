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

    util.wait_until_k8s_ready(cp1, [cp1])
    util.join_cluster(cp2, util.get_join_token(cp1, cp2))
    util.wait_until_k8s_ready(cp1, [cp1, cp2])
    assert util.get_datastore_type(cp1) == "etcd", "test needs the etcd datastore"
    _wait_for(
        lambda: sorted(_etcd_started_voters(cp1)) == sorted([cp1.id, cp2.id]),
        timeout_s=180,
        interval_s=5,
        what="cp1 and cp2 to be started etcd voters",
    )

    status = json.loads(
        cp1.exec(
            ["k8s", "status", "--output-format", "json"],
            capture_output=True,
            text=True,
        ).stdout
    )
    roles = {m["name"]: m["datastore-role"] for m in status["members"]}
    LOG.info("dqlite roles: %s", roles)
    spares = [cp for cp in (cp1, cp2) if roles.get(cp.id) == "spare"]
    assert len(spares) == 1, f"expected exactly one dqlite spare, got roles {roles}"
    spare = spares[0]
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
        "remove-node --force must fail while the node is still an etcd member "
        "and etcd has no quorum (canonical/k8sd#161)"
    )
    assert (
        "still an etcd member" in result.stderr
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
