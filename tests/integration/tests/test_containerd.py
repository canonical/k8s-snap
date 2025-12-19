#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.tags(tags.NIGHTLY)
def test_containerd(instances: List[harness.Instance]):
    instance = instances[0]
    util.wait_until_k8s_ready(instance, [instance])

    util.stubbornly(retries=5, delay_s=2).on(instance).exec(
        [
            "/snap/k8s/current/bin/ctr",
            "-n",
            "k8s.io",
            "images",
            "pull",
            "docker.io/library/nginx:1.29",
        ],
        capture_output=True,
        text=True,
        check=True,
    )

    # Test sideloading: export, delete, re-import and run image
    instance.exec(
        [
            "/snap/k8s/current/bin/ctr",
            "-n",
            "k8s.io",
            "image",
            "export",
            "/tmp/nginx-export.tar",
            "docker.io/library/nginx:1.29",
        ],
        capture_output=True,
        text=True,
        check=True,
    )

    instance.exec(
        [
            "/snap/k8s/current/bin/ctr",
            "-n",
            "k8s.io",
            "image",
            "rm",
            "docker.io/library/nginx:1.29",
        ],
        capture_output=True,
        text=True,
        check=True,
    )

    instance.exec(
        [
            "/snap/k8s/current/bin/ctr",
            "-n",
            "k8s.io",
            "image",
            "import",
            "/tmp/nginx-export.tar",
        ],
        capture_output=True,
        text=True,
        check=True,
    )

    result = instance.exec(
        ["/snap/k8s/current/bin/ctr", "-n", "k8s.io", "images", "ls"],
        capture_output=True,
        text=True,
    )
    assert "nginx:1.29" in result.stdout, (
        f"nginx image not found after sideloading\n"
        f"ctr output: {result.stdout}"
        f"ctr error: {result.stderr}"
    )

    # Run nginx pod with imagePullpolicy: never to ensure sideloaded image is used
    instance.exec(
        [
            "k8s",
            "kubectl",
            "run",
            "nginx-test",
            "--image=docker.io/library/nginx:1.29",
            "--image-pull-policy=Never",
            "-l",
            "run=my-nginx",
            "--port=80",
        ],
        capture_output=True,
        text=True,
        check=True,
    )

    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "wait",
            "--for=condition=ready",
            "pod",
            "-l",
            "run=my-nginx",
            "--timeout",
            "180s",
        ]
    )
