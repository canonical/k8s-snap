#
# Copyright 2026 Canonical, Ltd.
#
from typing import List

import pytest
from test_util import harness, tags, util


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_microk8s_installed(instances: List[harness.Instance]):
    instance = instances[0]
    util.stubbornly(retries=3, delay_s=30).on(instance).exec(
        "snap install microk8s --classic".split()
    )
    result = instance.exec("k8s bootstrap".split(), capture_output=True, check=False)
    assert "Error: microk8s snap is installed" in result.stderr.decode()

    instance.exec("snap remove microk8s --purge".split())
