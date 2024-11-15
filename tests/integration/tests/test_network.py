#
# Copyright 2024 Canonical, Ltd.
#
import json
import logging
from pathlib import Path
from typing import List

import pytest
from test_util import config, harness, util
from test_util.config import MANIFESTS_DIR

LOG = logging.getLogger(__name__)


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
def test_network(instances: List[harness.Instance]):
    instance = instances[0]
    util.wait_until_k8s_ready(instance, [instance])
    util.wait_for_network(instance)
    util.wait_for_dns(instance)

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
            "get",
            "pod",
            "-l",
            "app=nginx",
            "-o",
            "json",
        ],
        capture_output=True,
    )

    out = json.loads(p.stdout.decode())

    assert len(out["items"]) > 0, "No NGINX pod found"
    podIP = out["items"][0]["status"]["podIP"]

    util.stubbornly(retries=5, delay_s=5).on(instance).until(
        lambda p: "Welcome to nginx!" in p.stdout.decode()
    ).exec(["curl", "-s", f"http://{podIP}"])
