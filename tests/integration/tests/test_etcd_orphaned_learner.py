#
# Copyright 2026 Canonical, Ltd.
#
import json
import logging
from typing import List, Optional

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)

ETCD_PEER_PORT = 2380


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
    return _etcd_api(instance, "/v3/cluster/member/list").get("members", [])


def _join(cluster_node: harness.Instance, joining_node: harness.Instance):
    token = util.get_join_token(cluster_node, joining_node)
    result = joining_node.exec(
        ["k8s", "join-cluster", token], capture_output=True, text=True, check=False
    )
    if result.returncode != 0:
        members = [
            (m.get("name", ""), m.get("peerURLs"), m.get("isLearner", False))
            for m in _etcd_members(cluster_node)
        ]
        pytest.fail(
            f"join of {joining_node.id} failed: {result.stderr.strip()} - "
            f"etcd members: {members}"
        )


def _assert_all_named_voters(
    instance: harness.Instance, expected: List[harness.Instance]
):
    members = _etcd_members(instance)
    summary = [(m.get("name", ""), m.get("isLearner", False)) for m in members]
    assert sorted(m.get("name", "") for m in members) == sorted(
        i.id for i in expected
    ), f"unexpected etcd members: {summary}"
    assert not any(
        m.get("isLearner", False) for m in members
    ), f"learner left behind: {summary}"


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_join_recovers_orphaned_etcd_learner(instances: List[harness.Instance]):
    """A learner orphaned by a failed MemberAddAsLearner must not block joins.

    Regression test for canonical/k8sd#160. MemberAddAsLearner can return an
    error even though etcd committed the member (the response is lost, or the
    proposal commits after the client gave up). The join hook only learnt the
    learner ID from the response, so nothing cleaned it up. The result was an
    unnamed learner that blocked every later join, and the same node's retry
    failed with "Peer URLs already exists".

    The test creates exactly that state: an unnamed, never-started learner
    with cp2's peer URL. cp2's join must then adopt or replace it, and cp3
    must join afterwards without hitting etcd's one-learner limit.
    """
    cp1, cp2, cp3 = instances
    util.wait_until_k8s_ready(cp1, [cp1])

    cp2_peer_url = f"https://{util.get_default_ip(cp2)}:{ETCD_PEER_PORT}"
    added = _etcd_api(
        cp1, "/v3/cluster/member/add", {"peerURLs": [cp2_peer_url], "isLearner": True}
    )["member"]
    LOG.info("Created orphaned etcd learner %s for %s", added["ID"], cp2_peer_url)
    assert added.get("isLearner", False), "the orphan must be a learner"
    assert not added.get("name"), "a learner that never started has no name"

    _join(cp1, cp2)
    util.wait_until_k8s_ready(cp1, [cp1, cp2])
    _assert_all_named_voters(cp1, [cp1, cp2])

    _join(cp1, cp3)
    util.wait_until_k8s_ready(cp1, instances)
    _assert_all_named_voters(cp1, instances)

    util.stubbornly(retries=10, delay_s=3).until(
        lambda _: not util.diverged_cluster_memberships(cp1, instances)
    ).on(cp1).exec(["true"])
