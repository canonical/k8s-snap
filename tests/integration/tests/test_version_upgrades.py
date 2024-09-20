#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(3)
@pytest.mark.no_setup()
@pytest.mark.xfail("cilium failures are blocking this from working")
@pytest.mark.skipif(
    not config.VERSION_UPGRADE_CHANNELS, reason="No upgrade channels configured"
)
def test_version_upgrades(instances: List[harness.Instance]):
    channels = config.VERSION_UPGRADE_CHANNELS
    cp = instances[0]
    joining_cp = instances[1]
    worker = instances[2]

    # Setup the k8s snap from the bootstrap channel and setup basic configuration.
    cp.exec(["snap", "install", "k8s", "--channel", channels[0]])
    cp.exec(["k8s", "bootstrap"])

    # Create an initial cluster
    joining_cp.exec(["snap", "install", "k8s", "--channel", channels[0]])
    joining_cp_token = util.get_join_token(cp, joining_cp)
    joining_cp.exec(["k8s", "join-cluster", joining_cp_token])

    worker.exec(["snap", "install", "k8s", "--channel", channels[0]])
    worker_token = util.get_join_token(cp, worker, "--worker")
    worker.exec(["k8s", "join-cluster", worker_token])

    util.stubbornly(retries=30, delay_s=20).until(util.ready_nodes(cp) == 3)

    current_channel = channels[0]
    for channel in channels[1:]:
        for instance in instances:
            LOG.info(
                f"Upgrading {instance.id} from {current_channel} to channel {channel}"
            )
            # Log the current snap version on the node.
            instance.exec(["snap", "info", "k8s"])

            # note: the `--classic` flag will be ignored by snapd for strict snaps.
            instance.exec(
                ["snap", "refresh", "k8s", "--channel", channel, "--classic", "--amend"]
            )

            # After the refresh, do not wait until all nodes are up.
            # Microcluster expects other nodes to upgrade before continuing,
            # hence the node does not come up until all nodes are upgraded.

        # After each full upgrade, verify that all nodes are up (again)
        util.stubbornly(retries=30, delay_s=20).until(util.ready_nodes(cp) == 3)
        LOG.info(f"Upgraded {instance.id} to channel {channel}")
