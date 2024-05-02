#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
from pathlib import Path

from test_util import harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


def test_network(instance: harness.Instance):
    p = instance.exec(
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

    p = instance.exec(
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
    p = instance.exec(
        ["k8s", "kubectl", "apply", "-f", "-"],
        input=Path(manifest).read_bytes(),
    )

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
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
        ]
    )

    p = instance.exec(
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
