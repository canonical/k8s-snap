#
# Copyright 2026 Canonical, Ltd.
#
import concurrent.futures
from typing import List

import pytest
from test_util import config, harness, tags, util


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
def test_wrong_token_race(instances: List[harness.Instance]):
    cluster_node = instances[0]

    join_token = util.get_join_token(cluster_node, instances[1])
    util.join_cluster(instances[1], join_token)

    new_join_token = util.get_join_token(cluster_node, instances[2])

    util.wait_until_k8s_ready(cluster_node, instances[:2])

    # retry since the truststore entry can be populated to cluster_node before
    # removing the node, otherwise the node removal will fail with
    # "No truststore entry found for node"
    # The heartbeat is every 2 seconds, so waiting for 3 seconds
    # should be sufficient.
    util.remove_node_with_retry(cluster_node, instances[1].id, retries=3)

    another_join_token = util.get_join_token(cluster_node, instances[2])

    # The join token should have changed after the node was removed as
    # it contains the ip addresses of all cluster nodes.
    assert (
        new_join_token != another_join_token
    ), "join token is not updated after node removal"
    util.join_cluster(instances[2], new_join_token)


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.PULL_REQUEST, tags.NIGHTLY)
def test_concurrent_cp_join_race(
    instances: List[harness.Instance],
    registry,
    containerd_cfgdir: str,
):
    """Regression test for the concurrent control-plane join race
    (k8s-snap#2814, #2813).

    etcd allows only one pending (un-promoted) learner at a time, and
    after a promotion it requires a short (~5s) window of confirmed peer
    connectivity before it will accept the next membership change. Two
    control-plane nodes joining concurrently can therefore
    transiently hit etcd's ErrTooManyLearners or ErrUnhealthy on
    MemberAddAsLearner. k8sd should retry the join procedure when it
    encounters those errors.

    The concurrent join is repeated over several attempts so a pass does
    not rest on a single attempt not hitting the race. Between
    attempts the two joining nodes are removed from the cluster and their
    snap is purged and reinstalled, so each attempt runs against a fresh,
    unbootstrapped pair.
    """
    cluster_node = instances[0]
    joining_cp_A = instances[1]
    joining_cp_B = instances[2]
    util.wait_until_k8s_ready(cluster_node, [cluster_node])

    attempts = 5
    for attempt in range(attempts):
        join_token_A = util.get_join_token(cluster_node, joining_cp_A)
        join_token_B = util.get_join_token(cluster_node, joining_cp_B)

        with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
            future_A = executor.submit(util.join_cluster, joining_cp_A, join_token_A)
            future_B = executor.submit(util.join_cluster, joining_cp_B, join_token_B)
            for future in concurrent.futures.as_completed([future_A, future_B]):
                future.result()

        util.wait_until_k8s_ready(cluster_node, [joining_cp_A, joining_cp_B])
        assert "control-plane" in util.get_local_node_status(joining_cp_A)
        assert "control-plane" in util.get_local_node_status(joining_cp_B)

        if attempt == attempts - 1:
            break

        # Reset the joining pair for the next attempt.
        # Purging and reinstalling the snap wipes each node's local k8sd
        # state, so the next join starts from a genuinely fresh node.
        for node in (joining_cp_A, joining_cp_B):
            util.remove_node_with_retry(cluster_node, node.id)
        util.wait_until_k8s_ready(cluster_node, [cluster_node])

        for node in (joining_cp_A, joining_cp_B):
            util.remove_k8s_snap(node)
            util.setup_k8s_snap(node)
            # remove_k8s_snap wipes /etc/containerd (incl. hosts.d), so the
            # local registry mirror config must be re-applied after reinstall,
            # otherwise the next join pulls images directly and is subject to
            # registry rate limits and network connectivity issues.
            if config.USE_LOCAL_MIRROR:
                registry.apply_configuration(node, containerd_cfgdir)
