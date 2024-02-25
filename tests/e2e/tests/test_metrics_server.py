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
    util.wait_for_dns(instance)
    util.wait_for_network(instance)
    instance.exec(["k8s", "enable", "metrics-server"])

    LOG.info("Waiting for metrics-server pod to show up...")
    util.stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: "metrics-server" in p.stdout.decode()
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

    util.stubbornly(retries=15, delay_s=5).on(instance).until(
        lambda p: instance.id in p.stdout.decode()
    ).exec(["k8s", "kubectl", "top", "node"])
