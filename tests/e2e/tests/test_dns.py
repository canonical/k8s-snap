#
# Copyright 2024 Canonical, Ltd.
#
import logging
from pathlib import Path

import pytest
from e2e_util import config, harness, util

LOG = logging.getLogger(__name__)


def test_dns(h: harness.Harness, tmp_path: Path):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    snap_path = (tmp_path / "k8s.snap").as_posix()

    LOG.info("Creating instance")
    instance_id = h.new_instance()

    util.setup_k8s_snap(h, instance_id, snap_path)
    h.exec(instance_id, ["k8s", "bootstrap"])

    util.setup_network(h, instance_id)
    util.setup_dns(h, instance_id)

    h.exec(
        instance_id,
        [
            "k8s",
            "kubectl",
            "run",
            "busybox",
            "--image=busybox:1.28",
            "--restart=Never",
            "--",
            "sleep",
            "3600",
        ],
    )

    util.stubbornly(retries=3, delay_s=1).on(h, instance_id).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "run=busybox",
            "--timeout",
            "180s",
        ]
    )

    result = h.exec(
        instance_id,
        ["k8s", "kubectl", "exec", "busybox", "--", "nslookup", "kubernetes.default"],
        capture_output=True,
    )

    assert "10.152.183.1 kubernetes.default.svc.foo.local" in result.stdout.decode()

    result = h.exec(
        instance_id,
        ["k8s", "kubectl", "exec", "busybox", "--", "nslookup", "canonical.com"],
        capture_output=True,
        check=False,
    )

    assert "can't resolve" not in result.stdout.decode()

    h.cleanup()
