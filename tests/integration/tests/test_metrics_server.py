#
# Copyright 2024 Canonical, Ltd.
#
import logging

from test_util import harness, util

LOG = logging.getLogger(__name__)


def test_metrics_server(session_instance: harness.Instance):
    LOG.info("Waiting for metrics-server pod to show up...")
    util.stubbornly(retries=15, delay_s=5).on(session_instance).until(
        lambda p: "metrics-server" in p.stdout.decode()
    ).exec(["k8s", "kubectl", "get", "pod", "-n", "kube-system", "-o", "json"])
    LOG.info("Metrics-server pod showed up.")

    util.stubbornly(retries=3, delay_s=1).on(session_instance).exec(
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

    util.stubbornly(retries=15, delay_s=5).on(session_instance).until(
        lambda p: session_instance.id in p.stdout.decode()
    ).exec(["k8s", "kubectl", "top", "node"])
