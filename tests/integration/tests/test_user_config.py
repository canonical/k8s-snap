#
# Copyright 2025 Canonical, Ltd.
#
import json
import logging
from subprocess import CompletedProcess
from time import sleep
from typing import List

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)


@pytest.mark.node_count(1)
@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.NIGHTLY)
def test_no_unnecessary_helm_revisions(instances: List[harness.Instance]):
    """Verifies that calling 'k8s set' with the same configuration multiple times
    does not create new helm chart revisions.

    Tests that:
    - Multiple calls to 'k8s set' with the same values don't trigger unnecessary helm upgrades
    """
    instance = instances[0]

    # Bootstrap the cluster
    util.wait_until_k8s_ready(instance, [instance])

    def check_features_enabled(p: CompletedProcess):
        status = json.loads(p.stdout.strip())
        return (
            "enabled" in status.get("network", {}).get("message")
            and "enabled" in status.get("gateway", {}).get("message")
            and "enabled" in status.get("ingress", {}).get("message")
            and "enabled" in status.get("load-balancer", {}).get("message")
            and "enabled" in status.get("dns", {}).get("message")
            and "enabled" in status.get("local-storage", {}).get("message")
            and "enabled" in status.get("metrics-server", {}).get("message")
        )

    util.stubbornly(retries=50, delay_s=5).on(instance).until(
        check_features_enabled
    ).exec(
        ["k8s", "status", "--output-format", "json"],
        capture_output=True,
        text=True,
    )

    def get_chart_revisions():
        revisions = {}
        result = instance.exec(
            [
                "k8s",
                "helm",
                "list",
                "-A",
                "-o",
                "json",
            ],
            capture_output=True,
            text=True,
        )

        charts = json.loads(result.stdout)
        for chart in charts:
            revisions[chart["name"]] = chart["revision"]
        return revisions

    initial_revisions = get_chart_revisions()

    for i in range(10):
        LOG.info("Running k8s set command (iteration %d)", i + 1)
        instance.exec(
            [
                "k8s",
                "set",
                "network.enabled=true",
                "gateway.enabled=true",
                "ingress.enabled=true",
                "load-balancer.enabled=true",
                "dns.enabled=true",
                "local-storage.enabled=true",
                "metrics-server.enabled=true",
            ]
        )

        # NOTE(Hue): I can't think of a better way to make sure the features are reconciled.
        sleep(5)

        util.stubbornly(retries=50, delay_s=5).on(instance).until(
            check_features_enabled
        ).exec(
            ["k8s", "status", "--output-format", "json"],
            capture_output=True,
            text=True,
        )

        current_revisions = get_chart_revisions()

        assert len(current_revisions) == len(
            initial_revisions
        ), "Mismatch in number of charts"
        for chart, initial_rev in initial_revisions.items():
            current_rev = current_revisions.get(chart)
            assert current_rev is not None, f"Chart {chart} not in current revisions"
            if chart == "ck-dns":
                # NOTE(Hue): (KU-3683) ck-dns has a bug that clusterIP is not available in the first reconciliation
                # and release. And with the new reconciliation, the clusterIP is set and a new revision is
                # created. But this should only happen once. After that, the revision should not change.
                assert (
                    int(current_rev) == int(initial_rev) + 1
                ), f"Chart {chart} revision changed unexpectedly, expected {int(initial_rev) + 1}, got {current_rev}"
            else:
                assert (
                    current_rev == initial_rev
                ), f"Chart {chart} revision changed unexpectedly"
