#
# Copyright 2026 Canonical, Ltd.
#
"""Integration test for the LocalPV ConfigMap override feature.

This test verifies that the k8sd LocalPVConfigMapController picks up changes
to the `k8sd-localpv-values` ConfigMap in kube-system and applies them to
the LocalPV Helm release (ck-storage in kube-system).

Test flow:
1. Bootstrap a cluster with local-storage enabled, wait for it to be ready.
2. Apply an override ConfigMap with storageClass.reclaimPolicy=Delete
   (k8sd default is Retain).
3. Poll Helm values until the override is reflected.
4. Update the ConfigMap to storageClass.reclaimPolicy=Retain and verify.
5. Delete the ConfigMap and verify reclaimPolicy reverts to the k8sd default
   (Retain).
"""

import logging
from typing import List

import pytest
from test_util import config, configmap_override, harness, tags, util

LOG = logging.getLogger(__name__)

OVERRIDE_CM_NAME = "k8sd-localpv-values"
OVERRIDE_CM_NAMESPACE = "kube-system"
HELM_RELEASE = "ck-storage"
HELM_NAMESPACE = "kube-system"

# k8sd default for storageClass.reclaimPolicy.
DEFAULT_RECLAIM_POLICY = "Retain"


@pytest.mark.bootstrap_config((config.MANIFESTS_DIR / "bootstrap-all.yaml").read_text())
@pytest.mark.tags(tags.PULL_REQUEST)
def test_localpv_configmap_override(instances: List[harness.Instance]):
    """Verify that the LocalPVConfigMapController applies and reverts Helm overrides."""
    instance = instances[0]

    try:
        util.wait_until_k8s_ready(instance, [instance])

        # -- Step 1: Apply initial ConfigMap override --
        # Override storageClass.reclaimPolicy from the k8sd default (Retain) to Delete.
        LOG.info(
            "Applying LocalPV override ConfigMap with storageClass.reclaimPolicy=Delete"
        )
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "storageClass:\n  reclaimPolicy: Delete\n",
        )

        LOG.info("Waiting for Helm to reflect storageClass.reclaimPolicy=Delete")
        configmap_override.wait_for_override(
            instance,
            HELM_RELEASE,
            HELM_NAMESPACE,
            ["storageClass", "reclaimPolicy"],
            "Delete",
        )

        # -- Step 2: Update the ConfigMap override --
        # Retain is the other valid reclaimPolicy value (Recycle is deprecated).
        LOG.info(
            "Updating LocalPV override ConfigMap with storageClass.reclaimPolicy=Retain"
        )
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "storageClass:\n  reclaimPolicy: Retain\n",
        )

        LOG.info("Waiting for Helm to reflect storageClass.reclaimPolicy=Retain")
        configmap_override.wait_for_override(
            instance,
            HELM_RELEASE,
            HELM_NAMESPACE,
            ["storageClass", "reclaimPolicy"],
            "Retain",
        )

        # -- Step 3: Delete the ConfigMap and verify revert to k8sd defaults --
        LOG.info("Deleting LocalPV override ConfigMap")
        configmap_override.delete_override_configmap(
            instance, OVERRIDE_CM_NAME, OVERRIDE_CM_NAMESPACE
        )

        LOG.info(
            "Waiting for storageClass.reclaimPolicy to revert to k8sd default (%s)",
            DEFAULT_RECLAIM_POLICY,
        )
        configmap_override.wait_for_override(
            instance,
            HELM_RELEASE,
            HELM_NAMESPACE,
            ["storageClass", "reclaimPolicy"],
            DEFAULT_RECLAIM_POLICY,
        )

    finally:
        configmap_override.delete_override_configmap(
            instance, OVERRIDE_CM_NAME, OVERRIDE_CM_NAMESPACE
        )
