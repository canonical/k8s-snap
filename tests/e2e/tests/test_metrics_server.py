#
# Copyright 2024 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from e2e_util import config, harness, util

LOG = logging.getLogger(__name__)


def test_metrics_server(instances: List[harness.Instance]):
    if not config.SNAP:
        pytest.fail("Set TEST_SNAP to the path where the snap is")

    instance = instances[0]
    instance.exec(["k8s", "enable", "metrics-server"])

    LOG.info("Waiting for metrics-server pod to show up...")
    util.stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: "ck-metrics-server" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"])
    LOG.info("Metrics-server pod showed up.")

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
            "app.kubernetes.io/name=metrics-server",
            "--timeout",
            "180s",
        ]
    )

    instance.exec(["snap", "install", "go", "--channel=1.21/stable", "--classic"])
    instance.exec(
        [
            "git",
            "clone",
            "https://github.com/kubernetes-sigs/metrics-server.git",
            "--branch",
            "v0.7.0",
        ]
    )
    # Adjusting upstream e2e_tests for our deployment
    instance.exec(
        [
            "sed",
            "-i",
            "-e",
            "s#k8s-app=metrics-server#app.kubernetes.io/name=metrics-server#g",
            "metrics-server/test/e2e_test.go",
        ]
    )
    instance.exec(
        ["sed", "-i", "-e", "s#= 10250#= 10251#g", "metrics-server/test/e2e_test.go"]
    )

    instance.exec(["bash", "-c", "mkdir -p ~/.kube"])
    instance.exec(["bash", "-c", "k8s config > ~/.kube/config"])
    p = instance.exec(
        ["bash", "-c", "cd metrics-server && go test test/e2e_test.go -v -count=1"],
        capture_output=True,
    )
    LOG.info(p.stdout.decode())
