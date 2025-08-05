#
# Copyright 2025 Canonical, Ltd.
#
import logging
import re
import subprocess
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)

STATUS_PATTERNS = [
    r"cluster status:\s*ready",
    r"control plane nodes:\s*(\d{1,3}(?:\.\d{1,3}){3}:\d{1,5})\s\(voter\)",
    r"high availability:\s*no",
    r"network:\s*enabled",
    r"dns:\s*enabled at (\d{1,3}(?:\.\d{1,3}){3})",
    r"ingress:\s*enabled",
    r"load-balancer:\s*enabled, L2 mode",
    r"local-storage:\s*enabled at /var/snap/k8s/common/rawfile-storage",
    r"gateway\s*enabled",
]


@pytest.mark.tags(tags.PULL_REQUEST)
def test_feature_controller(instances: List[harness.Instance], datastore_type: str):
    """
    Verifies that the feature controller won't get stuck in a chaotic situation.
    """
    STATUS_PATTERNS.insert(3, r"datastore:\s*{}".format(datastore_type))

    instance = instances[0]

    for _ in range(20):
        instance.exec(
            [
                "k8s",
                "disable",
                "dns",
                "gateway",
                "ingress",
                "load-balancer",
                "metrics-server",
                "local-storage",
            ]
        )
        instance.exec(
            [
                "pkill",
                "-f",
                "-9",
                "/snap/k8s/current/bin/kube-apiserver",
            ],
            check=False,
        )
        instance.exec(
            [
                "k8s",
                "enable",
                "dns",
                "gateway",
                "ingress",
                "load-balancer",
                "metrics-server",
                "local-storage",
            ]
        )
        instance.exec(
            [
                "pkill",
                "-f",
                "-9",
                "/snap/k8s/current/bin/kube-apiserver",
            ],
            check=False,
        )

    def status_output_matches(p: subprocess.CompletedProcess) -> bool:
        result_lines = p.stdout.decode().strip().split("\n")
        if len(result_lines) != len(STATUS_PATTERNS):
            LOG.info(
                f"wrong number of results lines, expected {len(STATUS_PATTERNS)}, got {len(result_lines)}"
            )
            return False

        for i in range(len(result_lines)):
            line, pattern = result_lines[i], STATUS_PATTERNS[i]
            if not re.search(pattern, line):
                LOG.info(f"could not match `{line.strip()}` with `{pattern}`")
                return False

        return True

    LOG.info("Verifying the output of `k8s status`")
    util.stubbornly(retries=15, delay_s=10).on(instance).until(
        condition=status_output_matches,
    ).exec(["k8s", "status", "--wait-ready"])
