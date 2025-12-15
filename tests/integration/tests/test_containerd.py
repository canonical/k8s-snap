#
# Copyright 2025 Canonical, Ltd.
#
import logging
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.tags(tags.NIGHTLY)
def test_containerd(instances: List[harness.Instance]):
    instance = instances[0]
    util.wait_until_k8s_ready(instance, [instance])

    # Pull image
    result = instance.exec(
        [
            "/snap/k8s/current/bin/ctr",
            "-n",
            "k8s.io",
            "images",
            "pull",
            "docker.io/library/nginx:latest",
        ],
        capture_output=True,
        text=True,
    )
    assert result.returncode == 0, "Failed to pull nginx image"

    # Export, delete, re-import (test sideloading)
    result = instance.exec(
        [
            "/snap/k8s/current/bin/ctr",
            "-n",
            "k8s.io",
            "image",
            "export",
            "/tmp/nginx-export.tar",
            "docker.io/library/nginx:latest",
        ],
        capture_output=True,
        text=True,
    )
    assert result.returncode == 0, "Failed to export nginx image"

    result = instance.exec(
        [
            "/snap/k8s/current/bin/ctr",
            "-n",
            "k8s.io",
            "image",
            "rm",
            "docker.io/library/nginx:latest",
        ],
        capture_output=True,
        text=True,
    )
    assert result.returncode == 0, "Failed to remove nginx image"

    result = instance.exec(
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
    )
    assert result.returncode == 0, "Failed to import nginx image via sideloading"

    # Verify the image is available after sideloading
    result = instance.exec(
        ["/snap/k8s/current/bin/ctr", "-n", "k8s.io", "images", "ls"],
        capture_output=True,
        text=True,
    )
    assert "nginx:latest" in result.stdout, (
        f"nginx image not found after sideloading\n"
        f"ctr output: {result.stdout}"
        f"ctr error: {result.stderr}"
    )
