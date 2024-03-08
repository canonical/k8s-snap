#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import subprocess
from pathlib import Path
from typing import List

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


def test_storage(instances: List[harness.Instance]):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    instance = instances[0]
    instance.exec(["k8s", "enable", "local-storage"])

    LOG.info("Waiting for storage provisioner pod to show up...")
    util.stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: "ck-storage" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"])
    LOG.info("Storage provisioner pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
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

    manifest = MANIFESTS_DIR / "storage-setup.yaml"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for storage writer pod to show up...")
    util.stubbornly(retries=3, delay_s=10).on(instance).until(
        lambda p: "storage-writer-pod" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Storage writer pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
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
    util.stubbornly(retries=3, delay_s=1).on(instance).until(check_pvc_bound).exec(
        ["k8s", "kubectl", "get", "pvc", "-o", "json"]
    )
    LOG.info("Storage got provisioned and pvc is bound.")

    util.stubbornly(retries=5, delay_s=10).on(instance).until(
        lambda p: "LOREM IPSUM" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "logs", "storage-writer-pod"])

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "delete",
            "pod",
            "-l",
            "k8s-app=storage-writer-pod",
            "--force",
        ]
    )

    manifest = MANIFESTS_DIR / "storage-test.yaml"
    instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for storage reader pod to show up...")
    util.stubbornly(retries=3, delay_s=10).on(instance).until(
        lambda p: "storage-reader-pod" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Storage reader pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
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

    util.stubbornly(retries=5, delay_s=10).on(instance).until(
        lambda p: "LOREM IPSUM" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "logs", "storage-reader-pod"])

    LOG.info("Data can be read between pods.")
