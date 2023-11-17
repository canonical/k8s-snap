#
# Copyright 2023 Canonical, Ltd.
#
import logging
import subprocess
import time
from pathlib import Path

import config
import pytest
from conftest import Harness

LOG = logging.getLogger(__name__)


def test_smoke(h: Harness, tmp_path: Path):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    snap_path = (tmp_path / "k8s.snap").as_posix()

    LOG.info("Create instance")
    instance_id = h.new_instance()

    LOG.info("Install snap")
    h.send_file(instance_id, config.SNAP, snap_path)
    h.exec(instance_id, ["snap", "install", snap_path, "--dangerous"])

    LOG.info("Initialize Kubernetes")
    h.exec(instance_id, ["/snap/k8s/current/k8s/init.sh"])
    h.exec(instance_id, ["k8s", "init"])

    LOG.info("Start Kubernetes")
    h.exec(instance_id, ["k8s", "start"])

    success = False
    for attempt in range(30):
        try:
            LOG.info("(attempt %d) Waiting for Kubelet to register", attempt)
            p = h.exec(
                instance_id,
                ["k8s", "kubectl", "get", "node", instance_id, "--no-headers"],
                capture_output=True,
            )
            success = True
            LOG.info("Kubelet registered successfully!")
            p.stdout
            LOG.info("%s", p.stdout.decode())
            break
        except subprocess.CalledProcessError:
            time.sleep(5)

    if not success:
        pytest.fail("Kubelet node did not register")

    LOG.info("Remove Kubernetes")
    h.exec(instance_id, ["snap", "remove", "k8s", "--purge"])
