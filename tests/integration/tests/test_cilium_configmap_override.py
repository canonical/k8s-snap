#
# Copyright 2026 Canonical, Ltd.
#
"""Integration test for the Cilium ConfigMap override feature.

This test verifies that the k8sd CiliumConfigMapController picks up changes
to the `k8sd-cilium-values` ConfigMap in kube-system and applies them to
the Cilium Helm release (ck-network).

Test flow:
1. Bootstrap a cluster and wait for network to be ready.
2. Apply an override ConfigMap with a value k8sd does not set.
3. Poll Helm values until the override is reflected.
4. Update the ConfigMap and verify the new value is applied.
"""

import logging
from typing import List

import pytest
from test_util import config, configmap_override, harness, tags, util

LOG = logging.getLogger(__name__)

OVERRIDE_CM_NAME = "k8sd-cilium-values"
OVERRIDE_CM_NAMESPACE = "kube-system"
HELM_RELEASE = "ck-network"
HELM_NAMESPACE = "kube-system"


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_cilium_configmap_override(instances: List[harness.Instance]):
    """Verify that the CiliumConfigMapController applies and updates Helm overrides."""
    instance = instances[0]

    try:
        util.wait_until_k8s_ready(instance, [instance])

        # -- Step 1: Apply initial ConfigMap override --
        # Override bandwidthManager.enabled (k8sd does not set this, so it will
        # appear in helm get values only when we set it and disappear on delete).
        LOG.info(
            "Applying Cilium override ConfigMap with bandwidthManager.enabled=true"
        )
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "bandwidthManager:\n  enabled: true\n",
        )

        LOG.info("Waiting for Helm to reflect bandwidthManager.enabled=true")
        configmap_override.wait_for_override(
            instance,
            HELM_RELEASE,
            HELM_NAMESPACE,
            ["bandwidthManager", "enabled"],
            True,
        )

        # -- Step 2: Update the ConfigMap override --
        LOG.info(
            "Updating Cilium override ConfigMap with bandwidthManager.enabled=false"
        )
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "bandwidthManager:\n  enabled: false\n",
        )

        LOG.info("Waiting for Helm to reflect bandwidthManager.enabled=false")
        configmap_override.wait_for_override(
            instance,
            HELM_RELEASE,
            HELM_NAMESPACE,
            ["bandwidthManager", "enabled"],
            False,
        )

    finally:
        configmap_override.delete_override_configmap(
            instance, OVERRIDE_CM_NAME, OVERRIDE_CM_NAMESPACE
        )
