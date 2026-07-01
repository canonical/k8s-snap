#
# Copyright 2026 Canonical, Ltd.
#
"""Integration test for the metrics-server ConfigMap override feature.

This test verifies that the k8sd MetricsServerConfigMapController picks up
changes to the `k8sd-metrics-server-values` ConfigMap in kube-system and
applies them to the metrics-server Helm release (metrics-server in kube-system).

Test flow:
1. Bootstrap a cluster and wait for it to be ready.
2. Apply an override ConfigMap with a value k8sd does not set.
3. Poll Helm values until the override is reflected.
4. Update the ConfigMap and verify the new value is applied.
"""

import logging
from typing import List

import pytest
from test_util import config, configmap_override, harness, tags, util

LOG = logging.getLogger(__name__)

OVERRIDE_CM_NAME = "k8sd-metrics-server-values"
OVERRIDE_CM_NAMESPACE = "kube-system"
HELM_RELEASE = "metrics-server"
HELM_NAMESPACE = "kube-system"


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_metrics_server_configmap_override(instances: List[harness.Instance]):
    """Verify that the MetricsServerConfigMapController applies and updates Helm overrides."""
    instance = instances[0]

    try:
        util.wait_until_k8s_ready(instance, [instance])

        # -- Step 1: Apply initial ConfigMap override --
        # Override replicas (k8sd does not set this).
        LOG.info("Applying metrics-server override ConfigMap with replicas=2")
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "replicas: 2\n",
        )

        LOG.info("Waiting for Helm to reflect replicas=2")
        configmap_override.wait_for_override(
            instance, HELM_RELEASE, HELM_NAMESPACE, ["replicas"], 2
        )

        # -- Step 2: Update the ConfigMap override --
        LOG.info("Updating metrics-server override ConfigMap with replicas=3")
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "replicas: 3\n",
        )

        LOG.info("Waiting for Helm to reflect replicas=3")
        configmap_override.wait_for_override(
            instance, HELM_RELEASE, HELM_NAMESPACE, ["replicas"], 3
        )

    finally:
        configmap_override.delete_override_configmap(
            instance, OVERRIDE_CM_NAME, OVERRIDE_CM_NAMESPACE
        )
