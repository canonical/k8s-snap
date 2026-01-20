#
# Copyright 2026 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.tags(tags.NIGHTLY)
@pytest.mark.node_count(2)
def test_snap_services(instances: List[harness.Instance]):
    """
    Test that snap services are running after a `snap revert` instances.
    """

    cp = instances[0]
    worker = instances[1]
    token = util.get_join_token(cp, worker, "--worker")
    util.join_cluster(worker, token)

    LOG.info("Waiting for k8s to be ready")
    util.wait_until_k8s_ready(cp, instances)

    refresh_track = util.previous_track(config.SNAP)

    LOG.info(f"Refreshing the snap to {refresh_track}")
    cp.exec(f"snap refresh k8s --channel={refresh_track} --amend".split())
    worker.exec(f"snap refresh k8s --channel={refresh_track} --amend".split())

    LOG.info("Waiting for k8s to be ready")
    util.wait_until_k8s_ready(cp, instances)

    LOG.info("Reverting the snaps")
    cp.exec("snap revert k8s".split())
    worker.exec("snap revert k8s".split())

    LOG.info("Waiting for k8s to be ready")
    util.wait_until_k8s_ready(cp, instances)

    LOG.info("Checking snap services")
    util.check_snap_services_ready(cp, node_type="control-plane", datastore_type="etcd")
    util.check_snap_services_ready(worker, node_type="worker")
