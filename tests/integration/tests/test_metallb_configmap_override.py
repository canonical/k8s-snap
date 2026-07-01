#
# Copyright 2026 Canonical, Ltd.
#
"""Integration test for the MetalLB ConfigMap override feature.

This test verifies that the k8sd MetalLBConfigMapController picks up changes
to the `k8sd-metallb-values` ConfigMap in kube-system and applies them to
the MetalLB Helm release (metallb in metallb-system).

Test flow:
1. Bootstrap a cluster with load-balancer enabled, wait for it to be ready.
2. Apply an override ConfigMap with a value k8sd does not set.
3. Poll Helm values until the override is reflected.
4. Update the ConfigMap and verify the new value is applied.
"""

import logging
from typing import List

import pytest
from test_util import config, configmap_override, harness, tags, util

LOG = logging.getLogger(__name__)

OVERRIDE_CM_NAME = "k8sd-metallb-values"
OVERRIDE_CM_NAMESPACE = "kube-system"
HELM_RELEASE = "metallb"
HELM_NAMESPACE = "metallb-system"


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_metallb_configmap_override(instances: List[harness.Instance]):
    """Verify that the MetalLBConfigMapController applies and updates Helm overrides."""
    instance = instances[0]

    try:
        util.wait_until_k8s_ready(instance, [instance])

        # -- Step 1: Apply initial ConfigMap override --
        # Override controller.logLevel (k8sd does not set this value, so it will
        # appear in helm get values only when explicitly overridden).
        LOG.info("Applying MetalLB override ConfigMap with controller.logLevel=debug")
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "controller:\n  logLevel: debug\n",
        )

        LOG.info("Waiting for Helm to reflect controller.logLevel=debug")
        configmap_override.wait_for_override(
            instance, HELM_RELEASE, HELM_NAMESPACE, ["controller", "logLevel"], "debug"
        )

        # -- Step 2: Update the ConfigMap override --
        LOG.info("Updating MetalLB override ConfigMap with controller.logLevel=info")
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "controller:\n  logLevel: info\n",
        )

        LOG.info("Waiting for Helm to reflect controller.logLevel=info")
        configmap_override.wait_for_override(
            instance, HELM_RELEASE, HELM_NAMESPACE, ["controller", "logLevel"], "info"
        )

    finally:
        configmap_override.delete_override_configmap(
            instance, OVERRIDE_CM_NAME, OVERRIDE_CM_NAMESPACE
        )
