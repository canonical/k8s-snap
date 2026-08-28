#
# Copyright 2026 Canonical, Ltd.
#
import logging
import re
from typing import List

import pytest
from test_util import harness, tags, util

LOG = logging.getLogger(__name__)

# Injected into the sonobuoy manifest before running so that the e2e plugin
# produces verbose per-test output (required by CNCF for conformance submissions).
# sonobuoy's --plugin-env flag is unreliable in this environment, so we generate
# the manifest, patch it directly, then run from the patched file.
_INJECT_VERBOSE_SCRIPT = """\
content = open('sonobuoy_manifest.yaml').read()
patched = content.replace(
    '      - name: E2E_EXTRA_ARGS\\n',
    '      - name: E2E_EXTRA_GINKGO_ARGS\\n        value: --v\\n      - name: E2E_EXTRA_ARGS\\n',
    1,
)
assert '- name: E2E_EXTRA_GINKGO_ARGS' in patched, 'E2E_EXTRA_GINKGO_ARGS injection failed'
open('sonobuoy_manifest.yaml', 'w').write(patched)
"""


@pytest.mark.node_count(3)
@pytest.mark.tags(tags.CONFORMANCE)
def test_cncf_conformance(instances: List[harness.Instance]):
    cluster_node = cluster_setup(instances)
    install_sonobuoy(cluster_node)

    # Generate the conformance manifest so we can patch it directly.
    # sonobuoy gen auto-detects the k8s version from the running cluster.
    cluster_node.exec(
        [
            "./sonobuoy",
            "gen",
            "--plugin",
            "e2e",
            "--mode",
            "certified-conformance",
            ">",
            "sonobuoy_manifest.yaml",
        ],
    )

    # Inject E2E_EXTRA_GINKGO_ARGS=--v for verbose per-test output.
    cluster_node.exec(
        ["dd", "of=/tmp/inject_verbose.py"],
        input=_INJECT_VERBOSE_SCRIPT.encode(),
    )
    cluster_node.exec(["python3", "/tmp/inject_verbose.py"])

    # -f is incompatible with --wait, so start the run and wait separately.
    cluster_node.exec(["./sonobuoy", "run", "-f", "sonobuoy_manifest.yaml"])
    cluster_node.exec(["./sonobuoy", "wait", "--wait"])
    cluster_node.exec(
        ["./sonobuoy", "retrieve", "-f", "sonobuoy_e2e.tar.gz"],
    )
    cluster_node.exec(
        ["tar", "-xf", "sonobuoy_e2e.tar.gz", "--one-top-level"],
    )
    resp = cluster_node.exec(
        ["./sonobuoy", "results", "sonobuoy_e2e.tar.gz"],
        capture_output=True,
    )

    cluster_node.pull_file("/root/sonobuoy_e2e.tar.gz", "sonobuoy_e2e.tar.gz")

    output = resp.stdout.decode()
    LOG.info(output)
    failed_tests = int(re.search("Failed: (\\d+)", output).group(1))
    assert failed_tests == 0, f"{failed_tests} tests failed"


def cluster_setup(instances: List[harness.Instance]) -> harness.Instance:
    cluster_node = instances[0]
    joining_nodes = instances[1:]

    for joining_node in joining_nodes:
        join_token = util.get_join_token(cluster_node, joining_node)
        util.join_cluster(joining_node, join_token)

    util.wait_until_k8s_ready(cluster_node, instances)

    nodes = util.ready_nodes(cluster_node)
    assert len(nodes) == 3, "node should have joined cluster"
    assert "control-plane" in util.get_local_node_status(cluster_node)
    assert "control-plane" in util.get_local_node_status(joining_nodes[0])
    assert "control-plane" in util.get_local_node_status(joining_nodes[1])

    config = cluster_node.exec(["k8s", "config"], capture_output=True)
    cluster_node.exec(["dd", "of=/root/.kube/config"], input=config.stdout)

    return cluster_node


def install_sonobuoy(instance: harness.Instance):
    instance.exec(
        ["curl", "-L", util.sonobuoy_tar_gz(instance.arch), "-o", "sonobuoy.tar.gz"]
    )
    instance.exec(["tar", "xvzf", "sonobuoy.tar.gz"])
    instance.exec(["./sonobuoy", "version"])
