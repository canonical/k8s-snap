#
# Copyright 2023 Canonical, Ltd.
#
import logging
from pathlib import Path

import pytest
from e2e_util import config, harness, util

LOG = logging.getLogger(__name__)


def test_smoke(h: harness.Harness, tmp_path: Path):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    snap_path = (tmp_path / "k8s.snap").as_posix()

    LOG.info("Create instance")
    instance_id = h.new_instance()

    # TODO(bschimke): The node will not report ready as the CNI is not yet implemented in the k8s snap.
    #                 Set `wait_ready` to True once this is done.
    util.setup_k8s_snap(h, instance_id, snap_path, wait_ready=False)
