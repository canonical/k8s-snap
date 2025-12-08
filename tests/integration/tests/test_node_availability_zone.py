#
# Copyright 2025 Canonical, Ltd.
#
import hashlib
import logging
import struct
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)


def _get_failure_domain(availability_zone: str) -> int:
    # Generate sha256 hash, select the first 8 bytes and convert it
    # to a little endian uint64.
    hash_bytes = hashlib.sha256(bytes(availability_zone, "utf-8")).digest()
    return struct.unpack("<Q", hash_bytes[:8])[0]


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.parametrize("same_az", (False, True))
# For k8s-dqlite
@pytest.mark.required_ports(9000)
def test_node_availability_zone(
    instances: List[harness.Instance],
    same_az: bool,
    datastore_type: str,
):
    # Steps:
    # * create a three-node cluster
    # * set node availability zones
    #   * use the same AZ if "same_az" is set, unique AZs otherwise
    # * ensure that the right Dqlite failure domain is set based on the AZ
    # * ensure that the cluster is still available
    #   * k8sd and k8s-dqlite services are restarted in order for the failure
    #     domain changes to be applied. We need to make sure that this doesn't
    #     lead to a quorum loss.
    initial_node = instances[0]

    util.wait_until_k8s_ready(initial_node, [initial_node])

    joining_cplane_node_1 = instances[1]
    joining_cplane_node_2 = instances[2]

    join_token = util.get_join_token(initial_node, joining_cplane_node_1)
    join_token_2 = util.get_join_token(initial_node, joining_cplane_node_2)
    assert join_token != join_token_2

    util.join_cluster(joining_cplane_node_1, join_token)
    util.join_cluster(joining_cplane_node_2, join_token_2)

    util.wait_until_k8s_ready(initial_node, instances)

    def _get_az(instance, same_az, suffix):
        if same_az:
            return "fake-az"
        else:
            return instance.id

    # We're concerned about the risk of quorum loss after dqlite service
    # restarts. For this reason, we'll have multiple iterations, ensuring
    # that the cluster remains functional.
    iterations = 10
    for iteration in range(iterations):
        LOG.info("Starting iteration: %s", iteration)
        az_suffix = f"-{iteration}"

        # Apply the AZ labels.
        for instance in instances:
            az = _get_az(instance, same_az, az_suffix)
            util.stubbornly(retries=5, delay_s=10).on(instance).exec(
                [
                    "k8s",
                    "kubectl",
                    "label",
                    "nodes",
                    instance.id,
                    f"topology.kubernetes.io/zone={az}",
                    "--overwrite",
                ]
            )

        # Wait for the dqlite failure domain to be applied.
        for instance in instances:
            az = _get_az(instance, same_az, az_suffix)
            failure_domain = _get_failure_domain(az)

            LOG.info(
                "Node: %s, az: %s, expected failure domain: %s",
                instance.id,
                az,
                failure_domain,
            )
            util.stubbornly(retries=5, delay_s=10).on(instance).until(
                lambda p: str(failure_domain) in p.stdout.decode()
            ).exec(
                [
                    "cat",
                    "/var/snap/k8s/common/var/lib/k8sd/state/database/failure-domain",
                ]
            )

            if datastore_type == "k8s-dqlite":
                # Check k8s-dqlite.
                util.stubbornly(retries=5, delay_s=10).on(instance).until(
                    lambda p: str(failure_domain) in p.stdout.decode()
                ).exec(
                    ["cat", "/var/snap/k8s/common/var/lib/k8s-dqlite/failure-domain"]
                )

        # Make sure that the nodes remain available.
        util.wait_until_k8s_ready(initial_node, instances)
