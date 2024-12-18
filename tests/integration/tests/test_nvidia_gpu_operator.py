#
# Copyright 2025 Canonical, Ltd.
#

import logging
from typing import List, Mapping

import pytest
from test_util import config, harness, tags, util

LOG = logging.getLogger(__name__)

NVIDIA_GPU_OPERATOR_HELM_CHART_REPO = "https://helm.ngc.nvidia.com/nvidia"

# Mapping between the versions of the Nvidia `gpu-operator` and
# the host versions of Ubuntu they support.
# Because the `nvidia-driver-daemonset` pod included in the `gpu-operator`
# includes kernel drivers, its container image's release lifecycle is
# strictly tied to the version of Ubuntu on the host.
# https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/platform-support.html
NVIDIA_GPU_OPERATOR_SUPPORTED_UBUNTU_VERSIONS = {"v24.9.1": ["20.04", "22.04"]}

NVIDIA_KERNEL_MODULE_NAMES = ["nvidia", "nvidia_uvm", "nvidia_modeset"]

# Lifted 1:1 from:
# https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/getting-started.html#cuda-vectoradd
NVIDIA_CUDA_VECTOR_ADDITION_TEST_POD_NAME = "cuda-vectoradd"


def _check_nvidia_gpu_present(instance: harness.Instance) -> bool:
    """Checks whether at least one Nvidia GPU is available
    by exec-ing `lspci` on the target instance."""
    proc = instance.exec(["lspci", "-k"], capture_output=True, text=True)

    for line in proc.stdout.split("\n"):
        if "NVIDIA Corporation" in line:
            LOG.info(f"Found NVIDIA GPU in lspci output: {line}")
            return True

    LOG.info(f"Failed to find NVIDIA GPU in lspci output: {proc.stdout}")
    return False


def _check_nvidia_drivers_loaded(instance: harness.Instance) -> Mapping[str, bool]:
    """Ensures that Nvidia kernel modules are NOT loaded on
    the given harness instance."""

    proc = instance.exec(["lsmod"], capture_output=True, text=True)
    modules_present = {m: False for m in NVIDIA_KERNEL_MODULE_NAMES}
    for line in proc.stdout.split("\n"):
        for mod in modules_present:
            if line.startswith(mod):
                modules_present[mod] = True

    LOG.info(f"Located the following Nvidia kernel modules {modules_present}")
    return modules_present


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.WEEKLY)
@pytest.mark.tags(tags.GPU)
@pytest.mark.parametrize(
    "gpu_operator_version", NVIDIA_GPU_OPERATOR_SUPPORTED_UBUNTU_VERSIONS.keys()
)
def test_deploy_nvdia_gpu_operator(
    instances: List[harness.Instance], gpu_operator_version: str
):
    """Tests that the Nvidia `gpu-operator` can be deployed successfully
    using the upstream Helm chart and a sample application running a small
    CUDA workload gets scheduled and executed to completion.
    """
    instance = instances[0]
    test_namespace = "gpu-operator"

    # Prechecks to ensure the test instance is valid.
    if not _check_nvidia_gpu_present(instance):
        msg = (
            f"No Nvidia GPU present on harness instance '{instance.id}'. "
            "Skipping GPU-operator test."
        )
        LOG.warn(msg)
        pytest.skip(msg)

    # NOTE(aznashwan): considering the Nvidia gpu-operator's main purpose
    # is to set up the drivers on the nodes, and that running the `gpu-operator`
    # with pre-installed drivers can lead to incompatibilities between the
    # version of the drivers and the rest of the toolchain, we skip the test
    # if any of the drivers happened to be pre-loaded on the harness instance:
    modules_loaded = _check_nvidia_drivers_loaded(instance)
    if any(modules_loaded.values()):
        msg = (
            f"Cannot have any pre-loaded Nvidia GPU drivers before running "
            f"the Nvidia 'gpu-operator' test on instance {instance.id}. "
            f"Current Nvidia driver statuses: {modules_loaded}"
        )
        LOG.warn(msg)
        pytest.skip(msg)

    instance_release = util.get_os_version_id_for_instance(instance)
    if (
        instance_release
        not in NVIDIA_GPU_OPERATOR_SUPPORTED_UBUNTU_VERSIONS[gpu_operator_version]
    ):
        msg = (
            f"Unsupported Ubuntu release '{instance_release}' for `gpu-operator` "
            f"version '{gpu_operator_version}'. Skipping gpu-operator test."
        )
        LOG.warn(msg)
        pytest.skip(msg)

    # Add the upstream Nvidia GPU-operator Helm repo:
    instance.exec(
        ["k8s", "helm", "repo", "add", "nvidia", NVIDIA_GPU_OPERATOR_HELM_CHART_REPO]
    )
    instance.exec(["k8s", "helm", "repo", "update"])

    # Install `gpu-operator` chart:
    instance.exec(
        [
            "k8s",
            "helm",
            "install",
            "--generate-name",
            "--wait",
            "-n",
            test_namespace,
            "--create-namespace",
            "nvidia/gpu-operator",
            f"--version={gpu_operator_version}",
        ]
    )

    # Wait for the core daemonsets of the gpu-operator to be ready:
    daemonsets = [
        "nvidia-driver-daemonset",
        "nvidia-device-plugin-daemonset",
        "nvidia-container-toolkit-daemonset",
    ]
    # NOTE(aznashwan): it takes on average a little under 10 minutes for all
    # of the core daemonsets of the Nvidia GPU-operator to do their thing
    # on an AWS `g4dn.xlarge` instance (4 vCPUs/16GiB RAM), so we offer a
    # generous timeout of 15 minutes:
    for daemonset in daemonsets:
        util.wait_for_daemonset(
            instance,
            daemonset,
            namespace=test_namespace,
            retry_times=15,
            retry_delay_s=60,
        )

    # Deploy a sample CUDA app and let it run to completion:
    pod_spec_file = config.MANIFESTS_DIR / "cuda-vectoradd-nvidia-gpu-test-pod.yaml"
    pod_spec = pod_spec_file.read_text().format(
        NVIDIA_CUDA_VECTOR_ADDITION_TEST_POD_NAME
    )
    instance.exec(
        ["k8s", "kubectl", "-n", test_namespace, "apply", "-f", "-"],
        input=pod_spec.encode(),
    )
    util.stubbornly(retries=3, delay_s=1).on(instance).exec(
        [
            "k8s",
            "kubectl",
            "-n",
            test_namespace,
            "wait",
            "--for=condition=ready",
            "pod",
            NVIDIA_CUDA_VECTOR_ADDITION_TEST_POD_NAME,
            "--timeout",
            "180s",
        ]
    )
