#
# Copyright 2026 Canonical, Ltd.
#

import logging
from collections.abc import Mapping

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
NVIDIA_GPU_OPERATOR_SUPPORTED_UBUNTU_VERSIONS = {"v24.9.1": ["20.04", "22.04", "24.04"]}

NVIDIA_KERNEL_MODULE_NAMES = ["nvidia", "nvidia_uvm", "nvidia_modeset"]

# Lifted 1:1 from:
# https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/getting-started.html#cuda-vectoradd
NVIDIA_CUDA_VECTOR_ADDITION_TEST_POD_NAME = "cuda-vectoradd"


# PCI device classes that represent actual GPU hardware.
# Excludes PCI bridges, audio devices, and other non-GPU NVIDIA controllers
# (e.g. Jetson/Tegra SoC PCI bridges show "NVIDIA Corporation" but are not GPUs).
_NVIDIA_GPU_PCI_CLASSES = [
    "VGA compatible controller",
    "3D controller",
    "Display controller",
]


def _check_nvidia_gpu_present(instance: harness.Instance) -> bool:
    """Checks whether at least one discrete Nvidia GPU is available
    by exec-ing `lspci` on the target instance.

    Only matches actual GPU device classes (VGA, 3D controller), not PCI bridges
    or other NVIDIA controllers. The GPU Operator requires discrete GPUs —
    integrated GPUs (e.g. Jetson/Tegra) are not supported.
    """
    proc = instance.exec(["lspci", "-k"], capture_output=True, text=True)

    for line in proc.stdout.split("\n"):
        if "NVIDIA" in line and any(cls in line for cls in _NVIDIA_GPU_PCI_CLASSES):
            LOG.info(f"Found NVIDIA GPU in lspci output: {line}")
            return True

    LOG.info(
        "No discrete NVIDIA GPU found in lspci output. "
        "Jetson/Tegra integrated GPUs are not supported by the GPU Operator."
    )
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


def _dump_gpu_operator_diagnostics(instance: harness.Instance, namespace: str):
    """Dump operator-wide state for post-mortem debugging."""
    diagnostics = [
        (["k8s", "kubectl", "-n", namespace, "get", "pods", "-o", "wide"], "pods"),
        (
            ["k8s", "kubectl", "-n", namespace, "get", "daemonsets"],
            "daemonsets",
        ),
        (
            [
                "k8s",
                "kubectl",
                "-n",
                namespace,
                "get",
                "clusterpolicy",
                "-o",
                "yaml",
            ],
            "clusterpolicy",
        ),
        (
            [
                "k8s",
                "kubectl",
                "-n",
                namespace,
                "logs",
                "-l",
                "app=gpu-operator",
                "--tail=200",
            ],
            "gpu-operator controller logs",
        ),
        (
            [
                "k8s",
                "kubectl",
                "get",
                "events",
                "-n",
                namespace,
                "--sort-by=.lastTimestamp",
            ],
            "namespace events",
        ),
        (
            ["k8s", "kubectl", "get", "nodes", "-o", "yaml"],
            "node labels and status",
        ),
    ]
    for cmd, label in diagnostics:
        try:
            result = instance.exec(cmd, capture_output=True, text=True, check=False)
            LOG.warning("=== DIAG: %s ===\n%s", label, result.stdout)
            if result.stderr:
                LOG.warning("stderr: %s", result.stderr)
        except Exception as exc:
            LOG.warning("Failed to collect %s: %s", label, exc)

    _dump_failing_pod_logs(instance, namespace)


def _dump_failing_pod_logs(instance: harness.Instance, namespace: str):
    """Dump current+previous container logs for every non-Running pod."""
    try:
        proc = instance.exec(
            [
                "k8s",
                "kubectl",
                "-n",
                namespace,
                "get",
                "pods",
                "-o",
                "jsonpath={range .items[*]}{.metadata.name}|{.status.phase}\\n{end}",
            ],
            capture_output=True,
            text=True,
            check=False,
        )
    except Exception as exc:
        LOG.warning("Failed to list pods for log collection: %s", exc)
        return

    for line in (proc.stdout or "").strip().splitlines():
        if "|" not in line:
            continue
        pod_name, phase = line.split("|", 1)
        if phase.strip() == "Running":
            continue
        for flag, label in ((None, "current"), ("--previous", "previous")):
            cmd = [
                "k8s",
                "kubectl",
                "-n",
                namespace,
                "logs",
                pod_name,
                "--all-containers=true",
                "--tail=200",
            ]
            if flag:
                cmd.append(flag)
            try:
                r = instance.exec(cmd, capture_output=True, text=True, check=False)
                LOG.warning("=== DIAG: %s logs (%s) ===\n%s", pod_name, label, r.stdout)
                if r.stderr:
                    LOG.warning("stderr: %s", r.stderr)
            except Exception as exc:
                LOG.warning("Failed %s logs for %s: %s", label, pod_name, exc)


# Bootstrap YAML overrides cluster-config defaults, so we must re-enable the core
# features (network/dns/local-storage) alongside the containerd-base-dir override.
_GPU_BOOTSTRAP_CONFIG = (
    (
        f"containerd-base-dir: {config.CONTAINERD_BASE_DIR}\n"
        "cluster-config:\n"
        "  network:\n"
        "    enabled: true\n"
        "  dns:\n"
        "    enabled: true\n"
        "  local-storage:\n"
        "    enabled: true\n"
    )
    if config.CONTAINERD_BASE_DIR
    else None
)


@pytest.mark.node_count(1)
@pytest.mark.tags(tags.WEEKLY)
@pytest.mark.tags(tags.GPU)
@pytest.mark.bootstrap_config(_GPU_BOOTSTRAP_CONFIG)
@pytest.mark.parametrize(
    "gpu_operator_version", NVIDIA_GPU_OPERATOR_SUPPORTED_UBUNTU_VERSIONS.keys()
)
def test_deploy_nvidia_gpu_operator(
    instances: list[harness.Instance], gpu_operator_version: str
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
        LOG.warning(msg)
        pytest.skip(msg)

    # Check if drivers are already loaded on the instance.
    # When running inside LXD containers with GPU passthrough, the host's
    # kernel modules are visible. In this case, we tell the gpu-operator
    # to skip its driver installation and use the existing host drivers.
    modules_loaded = _check_nvidia_drivers_loaded(instance)
    host_drivers_present = any(modules_loaded.values())
    if host_drivers_present:
        LOG.info(
            "Nvidia drivers already loaded on instance '%s'. "
            "Will deploy gpu-operator with driver.enabled=false. "
            "Driver statuses: %s",
            instance.id,
            modules_loaded,
        )

    instance_release = util.get_os_version_id_for_instance(instance)
    if (
        instance_release
        not in NVIDIA_GPU_OPERATOR_SUPPORTED_UBUNTU_VERSIONS[gpu_operator_version]
    ):
        msg = (
            f"Unsupported Ubuntu release '{instance_release}' for `gpu-operator` "
            f"version '{gpu_operator_version}'. Skipping gpu-operator test."
        )
        LOG.warning(msg)
        pytest.skip(msg)

    LOG.info("Waiting for k8s node to become Ready (CNI initialized)...")
    util.wait_until_k8s_ready(instance, instances)

    if config.CONTAINERD_BASE_DIR:
        # gpu-operator hard-codes hostPath volume mounts at /etc/containerd and
        # /run/containerd; bind-mount so the defaults reach our relocated paths.
        for target, source in (
            ("/etc/containerd", f"{config.CONTAINERD_BASE_DIR}/etc/containerd"),
            ("/run/containerd", f"{config.CONTAINERD_BASE_DIR}/run/containerd"),
        ):
            instance.exec(["mkdir", "-p", target])
            instance.exec(
                [
                    "bash",
                    "-c",
                    f"mountpoint -q {target} || mount --bind {source} {target}",
                ]
            )

    # Add the upstream Nvidia GPU-operator Helm repo:
    instance.exec(
        ["k8s", "helm", "repo", "add", "nvidia", NVIDIA_GPU_OPERATOR_HELM_CHART_REPO]
    )
    instance.exec(["k8s", "helm", "repo", "update"])

    # Install `gpu-operator` chart:
    helm_install_cmd = [
        "k8s",
        "helm",
        "install",
        "--generate-name",
        "-n",
        test_namespace,
        "--create-namespace",
        "nvidia/gpu-operator",
        f"--version={gpu_operator_version}",
    ]
    if host_drivers_present:
        helm_install_cmd.append("--set=driver.enabled=false")

    instance.exec(helm_install_cmd)

    # Wait for the core daemonsets of the gpu-operator to be ready:
    daemonsets = [
        "nvidia-device-plugin-daemonset",
        "nvidia-container-toolkit-daemonset",
    ]
    if not host_drivers_present:
        daemonsets.insert(0, "nvidia-driver-daemonset")
    # NOTE(aznashwan): it takes on average a little under 10 minutes for all
    # of the core daemonsets of the Nvidia GPU-operator to do their thing
    # on an AWS `g4dn.xlarge` instance (4 vCPUs/16GiB RAM), so we offer a
    # generous timeout of 15 minutes:
    for daemonset in daemonsets:
        try:
            util.wait_for_daemonset(
                instance,
                daemonset,
                namespace=test_namespace,
                retry_times=15,
                retry_delay_s=60,
            )
        except AssertionError:
            LOG.warning(
                "Daemonset '%s' never became ready — collecting diagnostics",
                daemonset,
            )
            _dump_gpu_operator_diagnostics(instance, test_namespace)
            raise

    # Wait for nvidia.com/gpu resources to be advertised on the node.
    # The device-plugin may be "Ready" but not yet registered GPU resources.
    LOG.info("Waiting for nvidia.com/gpu resources to appear on the node...")
    util.stubbornly(retries=30, delay_s=10).on(instance).until(
        lambda p: "nvidia.com/gpu" in p.stdout.decode()
    ).exec(
        [
            "k8s",
            "kubectl",
            "get",
            "nodes",
            "-o",
            "jsonpath={.items[*].status.allocatable}",
        ],
        capture_output=True,
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
    try:
        util.stubbornly(retries=5, delay_s=1).on(instance).exec(
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
    except Exception:
        LOG.warning("CUDA pod never became ready — collecting diagnostics")
        try:
            result = instance.exec(
                [
                    "k8s",
                    "kubectl",
                    "-n",
                    test_namespace,
                    "describe",
                    "pod",
                    NVIDIA_CUDA_VECTOR_ADDITION_TEST_POD_NAME,
                ],
                capture_output=True,
                text=True,
                check=False,
            )
            LOG.warning("=== DIAG: cuda-vectoradd describe ===\n%s", result.stdout)
        except Exception as diag_exc:
            LOG.warning("Failed to describe CUDA pod: %s", diag_exc)
        _dump_gpu_operator_diagnostics(instance, test_namespace)
        raise
