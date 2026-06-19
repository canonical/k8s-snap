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
5. Delete the ConfigMap and verify the override is absent from Helm values.
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


def _wait_for_replicas(
    instance: harness.Instance,
    expected: int,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until replicas matches expected."""

    def _replicas_match(p) -> bool:
        values = configmap_override.parse_helm_stdout(p)
        actual = values.get("replicas")
        LOG.info("Helm replicas: %s (want %s)", actual, expected)
        return actual == expected

    util.stubbornly(retries=retries, delay_s=delay_s).on(instance).until(
        _replicas_match
    ).exec(
        configmap_override.helm_values_cmd(HELM_RELEASE, HELM_NAMESPACE),
    )


def _wait_for_replicas_absent(
    instance: harness.Instance,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until replicas is absent (override reverted)."""

    def _replicas_absent(p) -> bool:
        values = configmap_override.parse_helm_stdout(p)
        absent = "replicas" not in values
        LOG.info("Helm values 'replicas' absent: %s (want True)", absent)
        return absent

    util.stubbornly(retries=retries, delay_s=delay_s).on(instance).until(
        _replicas_absent
    ).exec(
        configmap_override.helm_values_cmd(HELM_RELEASE, HELM_NAMESPACE),
    )


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_metrics_server_configmap_override(instances: List[harness.Instance]):
    """Verify that the MetricsServerConfigMapController applies and reverts Helm overrides."""
    instance = instances[0]

    try:
        util.wait_until_k8s_ready(instance, [instance])

        # -- Step 1: Apply initial ConfigMap override --
        # Override replicas (k8sd does not set this, so it will appear in
        # helm get values only when explicitly overridden and disappear on delete).
        LOG.info("Applying metrics-server override ConfigMap with replicas=2")
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "replicas: 2\n",
        )

        LOG.info("Waiting for Helm to reflect replicas=2")
        _wait_for_replicas(instance, expected=2)

        # -- Step 2: Update the ConfigMap override --
        LOG.info("Updating metrics-server override ConfigMap with replicas=3")
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "replicas: 3\n",
        )

        LOG.info("Waiting for Helm to reflect replicas=3")
        _wait_for_replicas(instance, expected=3)

        # -- Step 3: Delete the ConfigMap and verify revert --
        LOG.info("Deleting metrics-server override ConfigMap")
        configmap_override.delete_override_configmap(
            instance, OVERRIDE_CM_NAME, OVERRIDE_CM_NAMESPACE
        )

        LOG.info("Waiting for 'replicas' to be absent from Helm values")
        _wait_for_replicas_absent(instance)

    finally:
        configmap_override.delete_override_configmap(
            instance, OVERRIDE_CM_NAME, OVERRIDE_CM_NAMESPACE
        )
