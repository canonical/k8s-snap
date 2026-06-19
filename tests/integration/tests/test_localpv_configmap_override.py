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
4. Update the ConfigMap and verify the new value is applied.
5. Delete the ConfigMap and verify reclaimPolicy reverts to the k8sd default
   (Retain).
"""

import logging
from typing import List

import pytest
from test_util import config, harness, tags, util, configmap_override

LOG = logging.getLogger(__name__)

OVERRIDE_CM_NAME = "k8sd-localpv-values"
OVERRIDE_CM_NAMESPACE = "kube-system"
HELM_RELEASE = "ck-storage"
HELM_NAMESPACE = "kube-system"

# k8sd default for storageClass.reclaimPolicy.
DEFAULT_RECLAIM_POLICY = "Retain"


def _wait_for_reclaim_policy(
    instance: harness.Instance,
    expected: str,
    retries: int = 30,
    delay_s: int = 5,
):
    """Poll Helm values until storageClass.reclaimPolicy matches expected."""

    def _policy_matches(p) -> bool:
        values = configmap_override.parse_helm_stdout(p)
        actual = values.get("storageClass", {}).get("reclaimPolicy")
        LOG.info("Helm storageClass.reclaimPolicy: %s (want %s)", actual, expected)
        return actual == expected

    util.stubbornly(retries=retries, delay_s=delay_s).on(instance).until(
        _policy_matches
    ).exec(
        configmap_override.helm_values_cmd(HELM_RELEASE, HELM_NAMESPACE),
    )


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
        _wait_for_reclaim_policy(instance, expected="Delete")

        # -- Step 2: Update the ConfigMap override --
        LOG.info(
            "Updating LocalPV override ConfigMap with storageClass.reclaimPolicy=Recycle"
        )
        configmap_override.apply_override_configmap(
            instance,
            OVERRIDE_CM_NAME,
            OVERRIDE_CM_NAMESPACE,
            "storageClass:\n  reclaimPolicy: Recycle\n",
        )

        LOG.info("Waiting for Helm to reflect storageClass.reclaimPolicy=Recycle")
        _wait_for_reclaim_policy(instance, expected="Recycle")

        # -- Step 3: Delete the ConfigMap and verify revert to k8sd defaults --
        LOG.info("Deleting LocalPV override ConfigMap")
        configmap_override.delete_override_configmap(
            instance, OVERRIDE_CM_NAME, OVERRIDE_CM_NAMESPACE
        )

        LOG.info(
            "Waiting for storageClass.reclaimPolicy to revert to k8sd default (%s)",
            DEFAULT_RECLAIM_POLICY,
        )
        _wait_for_reclaim_policy(instance, expected=DEFAULT_RECLAIM_POLICY)

    finally:
        configmap_override.delete_override_configmap(
            instance, OVERRIDE_CM_NAME, OVERRIDE_CM_NAMESPACE
        )
