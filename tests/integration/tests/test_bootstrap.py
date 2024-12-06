#
# Copyright 2024 Canonical, Ltd.
#
from typing import List

import pytest
from test_util import config, harness, tags


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_microk8s_installed(instances: List[harness.Instance]):
    instance = instances[0]
    instance.exec("snap install microk8s --classic".split())
    result = instance.exec("k8s bootstrap".split(), capture_output=True, check=False)
    assert "Error: microk8s snap is installed" in result.stderr.decode()


@pytest.mark.node_count(1)
@pytest.mark.disable_k8s_bootstrapping()
@pytest.mark.tags(tags.NIGHTLY)
def test_k8s_service_invalid_extra_args(instances: List[harness.Instance]):
    instance = instances[0]
    foo_lish_bootstrap_config = (
        config.MANIFESTS_DIR / "bootstrap-foo-lish-arg.yaml"
    ).read_text()

    result = instance.exec(
        ["k8s", "bootstrap", "--file", "-"],
        input=str.encode(foo_lish_bootstrap_config),
        capture_output=True,
        text=True,
        check=False,
    )

    assert "expected kubelet to be in state active" in result.stderr
