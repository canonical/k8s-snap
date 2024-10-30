#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
import subprocess
from pathlib import Path

from test_util import harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def check_pvc_bound(p: subprocess.CompletedProcess) -> bool:
    out = json.loads(p.stdout.decode())
    for pvc in out["items"]:
        if pvc["metadata"]["name"] == "myclaim" and pvc["status"]["phase"] == "Bound":
            return True
    return False


def test_storage(aio_instance: harness.Instance):
    LOG.info("Waiting for storage provisioner pod to show up...")
    util.stubbornly(retries=15, delay_s=5).on(aio_instance).until(
        lambda p: "ck-storage" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"])
    LOG.info("Storage provisioner pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(aio_instance).exec(
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
    aio_instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for storage writer pod to show up...")
    util.stubbornly(retries=3, delay_s=10).on(aio_instance).until(
        lambda p: "storage-writer-pod" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Storage writer pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(aio_instance).exec(
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
    util.stubbornly(retries=3, delay_s=1).on(aio_instance).until(check_pvc_bound).exec(
        ["k8s", "kubectl", "get", "pvc", "-o", "json"]
    )
    LOG.info("Storage got provisioned and pvc is bound.")

    util.stubbornly(retries=5, delay_s=10).on(aio_instance).until(
        lambda p: "LOREM IPSUM" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "logs", "storage-writer-pod"])

    util.stubbornly(retries=3, delay_s=1).on(aio_instance).exec(
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
    aio_instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    LOG.info("Waiting for storage reader pod to show up...")
    util.stubbornly(retries=3, delay_s=10).on(aio_instance).until(
        lambda p: "storage-reader-pod" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-o", "json"])
    LOG.info("Storage reader pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(aio_instance).exec(
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

    util.stubbornly(retries=5, delay_s=10).on(aio_instance).until(
        lambda p: "LOREM IPSUM" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "logs", "storage-reader-pod"])

    LOG.info("Data can be read between pods.")
