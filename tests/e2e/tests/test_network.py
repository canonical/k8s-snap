#
# Copyright 2023 Canonical, Ltd.
#
import json
import logging
from pathlib import Path

import pytest
from e2e_util import config, harness, util
from e2e_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def test_network(h: harness.Harness, tmp_path: Path):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    snap_path = (tmp_path / "k8s.snap").as_posix()

    LOG.info("Create instance")
    instance_id = h.new_instance()

    util.setup_k8s_snap(h, instance_id, snap_path)
    h.exec(instance_id, ["k8s", "init"])
    util.setup_network(h, instance_id)

    p = h.exec(
        instance_id,
        [
            "k8s",
            "kubectl",
            "get",
            "pod",
            "-n",
            "kube-system",
            "-l",
            "k8s-app=cilium",
            "-o",
            "json",
        ],
        capture_output=True,
    )

    out = json.loads(p.stdout.decode())
    assert len(out["items"]) > 0

    cilium_pod = out["items"][0]

    p = h.exec(
        instance_id,
        [
            "k8s",
            "kubectl",
            "exec",
            "-it",
            cilium_pod["metadata"]["name"],
            "-n",
            "kube-system",
            "-c",
            "cilium-agent",
            "--",
            "cilium",
            "status",
            "--brief",
        ],
        capture_output=True,
    )

    assert p.stdout.decode().strip() == "OK"

    manifest = MANIFESTS_DIR / "nginx-pod.yaml"
    p = h.exec(
        instance_id,
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    util.retry_until_condition(
        h,
        instance_id,
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "app=nginx",
            "--timeout",
            "180s",
        ],
        max_retries=3,
        delay_between_retries=1,
    )

    p = h.exec(
        instance_id,
        [
            "k8s",
            "kubectl",
            "exec",
            "-it",
            cilium_pod["metadata"]["name"],
            "-n",
            "kube-system",
            "-c",
            "cilium-agent",
            "--",
            "cilium",
            "endpoint",
            "list",
            "-o",
            "json",
        ],
        capture_output=True,
    )
    assert "nginx" in p.stdout.decode().strip()

    h.cleanup()
