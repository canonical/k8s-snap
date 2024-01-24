#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import subprocess
from pathlib import Path

import pytest
from e2e_util import config, harness, util
from e2e_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def check_pvc_bound(p: subprocess.CompletedProcess) -> bool:
    out = json.loads(p.stdout.decode())
    for pvc in out["items"]:
        if pvc["metadata"]["name"] == "myclaim" and pvc["status"]["phase"] == "Bound":
            return True
    return False


def test_storage(h: harness.Harness, tmp_path: Path):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    snap_path = (tmp_path / "k8s.snap").as_posix()

    LOG.info("Create instance")
    instance_id = h.new_instance()

    util.setup_k8s_snap(h, instance_id, snap_path)
    h.exec(instance_id, ["k8s", "bootstrap"])
    util.setup_network(h, instance_id)

    out = h.exec(
        instance_id,
        ["k8s", "enable", "storage"],
        capture_output=True,
    )
    assert out.returncode == 0

    LOG.info("Waiting for storage provisioner pod to show up...")
    util.stubbornly(retries=15, delay_s=5).on(h, instance_id).until(
        lambda p: "ck-storage" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"])
    LOG.info("Storage provisioner pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(h, instance_id).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-n",
            "kube-system",
            "-l",
            "app.kubernetes.io/instance=ck-storage",
            "--timeout",
            "180s",
        ]
    )

    manifest = MANIFESTS_DIR / "storage-test.yaml"
    h.exec(
        instance_id,
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for storage writer pod to show up...")
    util.stubbornly(retries=3, delay_s=10).on(h, instance_id).until(
        lambda p: "storage-writer-pod" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Storage writer pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(h, instance_id).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "k8s-app=storage-writer-pod",
            "--timeout",
            "180s",
        ]
    )

    LOG.info("Waiting for storage to get provisioned...")
    util.stubbornly(retries=3, delay_s=1).on(h, instance_id).until(
        check_pvc_bound
    ).exec(["k8s", "kubectl", "get", "pvc", "-o", "json"])
    LOG.info("Storage got provisioned and pvc is bound.")

    LOG.info("Waiting for storage reader pod to show up...")
    util.stubbornly(retries=3, delay_s=10).on(h, instance_id).until(
        lambda p: "storage-reader-pod" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Storage reader pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(h, instance_id).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "k8s-app=storage-reader-pod",
            "--timeout",
            "180s",
        ]
    )

    util.stubbornly(retries=5, delay_s=10).on(h, instance_id).until(
        lambda p: "LOREM IPSUM" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "logs", "storage-reader-pod"])

    LOG.info("Data can be read between pods.")

    h.cleanup()
