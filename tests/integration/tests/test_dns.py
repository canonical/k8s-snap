#
# Copyright 2024 Canonical, Ltd.
#
import logging

from test_util import harness, util

LOG = logging.getLogger(__name__)


def test_dns(aio_instance: harness.Instance):
    aio_instance.exec(
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

    util.stubbornly(retries=3, delay_s=1).on(aio_instance).exec(
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

    result = aio_instance.exec(
        ["k8s", "kubectl", "exec", "busybox", "--", "nslookup", "kubernetes.default"],
        capture_output=True,
    )

    assert "10.152.183.1 kubernetes.default.svc.cluster.local" in result.stdout.decode()

    result = aio_instance.exec(
        ["k8s", "kubectl", "exec", "busybox", "--", "nslookup", "canonical.com"],
        capture_output=True,
        check=False,
    )

    assert "can't resolve" not in result.stdout.decode()
